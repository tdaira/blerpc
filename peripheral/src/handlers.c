#include "handlers.h"
#include "ble_service.h"
#include "blerpc.pb.h"
#include <blerpc_protocol/command.h>
#include <blerpc_protocol/container.h>
#include <pb_encode.h>
#include <pb_decode.h>
#include <string.h>
#include <zephyr/kernel.h>
#include <zephyr/logging/log.h>
#include <zephyr/drivers/flash.h>

LOG_MODULE_REGISTER(handlers, LOG_LEVEL_INF);

#define MAX_FLASH_READ_SIZE 8192

int handle_echo(const uint8_t *req_data, size_t req_len, pb_ostream_t *ostream)
{
    blerpc_EchoRequest req = blerpc_EchoRequest_init_zero;
    pb_istream_t stream = pb_istream_from_buffer(req_data, req_len);

    if (!pb_decode(&stream, blerpc_EchoRequest_fields, &req)) {
        LOG_ERR("Echo decode failed: %s", PB_GET_ERROR(&stream));
        return -1;
    }

    LOG_INF("Echo: \"%s\"", req.message);

    blerpc_EchoResponse resp = blerpc_EchoResponse_init_zero;
    strncpy(resp.message, req.message, sizeof(resp.message) - 1);

    if (!pb_encode(ostream, blerpc_EchoResponse_fields, &resp)) {
        LOG_ERR("Echo encode failed: %s", PB_GET_ERROR(ostream));
        return -1;
    }

    return 0;
}

struct flash_encode_ctx {
    const struct device *flash_dev;
    uint32_t address;
    uint32_t length;
};

static bool flash_data_encode_cb(pb_ostream_t *stream, const pb_field_t *field, void *const *arg)
{
    struct flash_encode_ctx *ctx = *(struct flash_encode_ctx **)arg;

    if (!pb_encode_tag_for_field(stream, field))
        return false;
    if (!pb_encode_varint(stream, ctx->length))
        return false;

    /* Read flash in chunks and stream directly to protobuf encoder */
    uint8_t chunk[256];
    uint32_t addr = ctx->address;
    uint32_t remaining = ctx->length;
    while (remaining > 0) {
        uint32_t n = MIN(remaining, sizeof(chunk));
        if (flash_read(ctx->flash_dev, addr, chunk, n) != 0)
            return false;
        if (!pb_write(stream, chunk, n))
            return false;
        addr += n;
        remaining -= n;
    }
    return true;
}

int handle_flash_read(const uint8_t *req_data, size_t req_len, pb_ostream_t *ostream)
{
    blerpc_FlashReadRequest req = blerpc_FlashReadRequest_init_zero;
    pb_istream_t stream = pb_istream_from_buffer(req_data, req_len);

    if (!pb_decode(&stream, blerpc_FlashReadRequest_fields, &req)) {
        LOG_ERR("FlashRead decode failed: %s", PB_GET_ERROR(&stream));
        return -1;
    }

    LOG_INF("FlashRead: addr=0x%08x len=%u", req.address, req.length);

    if (req.length > MAX_FLASH_READ_SIZE) {
        LOG_ERR("FlashRead: requested length %u exceeds max %d", req.length, MAX_FLASH_READ_SIZE);
        return -1;
    }

    const struct device *flash_dev = DEVICE_DT_GET(DT_CHOSEN(zephyr_flash_controller));
    if (!device_is_ready(flash_dev)) {
        LOG_ERR("Flash device not ready");
        return -1;
    }

    struct flash_encode_ctx ctx = {
        .flash_dev = flash_dev,
        .address = req.address,
        .length = req.length,
    };

    blerpc_FlashReadResponse resp = blerpc_FlashReadResponse_init_zero;
    resp.address = req.address;
    resp.data.funcs.encode = flash_data_encode_cb;
    resp.data.arg = &ctx;

    if (!pb_encode(ostream, blerpc_FlashReadResponse_fields, &resp)) {
        LOG_ERR("FlashRead encode failed: %s", PB_GET_ERROR(ostream));
        return -1;
    }

    return 0;
}

/* Callback for decoding DataWriteRequest.data — count bytes, discard data */
struct data_write_decode_ctx {
    uint32_t total_bytes;
};

static bool data_write_decode_cb(pb_istream_t *stream, const pb_field_t *field, void **arg)
{
    (void)field;
    struct data_write_decode_ctx *ctx = (struct data_write_decode_ctx *)*arg;

    size_t len = stream->bytes_left;
    ctx->total_bytes += (uint32_t)len;

    /* Discard data by reading into a small buffer */
    uint8_t discard[256];
    while (len > 0) {
        size_t n = MIN(len, sizeof(discard));
        if (!pb_read(stream, discard, n)) {
            return false;
        }
        len -= n;
    }
    return true;
}

/* ── counter_stream: P→C stream ───────────────────────────────────── */

static int notify_send_cb(const uint8_t *data, size_t len, void *ctx)
{
    (void)ctx;
    int rc;
    for (int retries = 0; retries < 10; retries++) {
        rc = ble_service_notify(data, len);
        if (rc != -ENOMEM) {
            return rc;
        }
        k_sleep(K_MSEC(5));
    }
    LOG_ERR("Notify send failed after retries: %d", rc);
    return rc;
}

static int send_one_counter_stream_response(uint32_t seq, int32_t value)
{
    blerpc_CounterStreamResponse resp = blerpc_CounterStreamResponse_init_zero;
    resp.seq = seq;
    resp.value = value;

    /* Encode protobuf to buffer */
    uint8_t pb_buf[blerpc_CounterStreamResponse_size];
    pb_ostream_t ostream = pb_ostream_from_buffer(pb_buf, sizeof(pb_buf));
    if (!pb_encode(&ostream, blerpc_CounterStreamResponse_fields, &resp)) {
        LOG_ERR("CounterStream encode failed");
        return -1;
    }

    /* Build command response */
    static uint8_t cmd_buf[64];
    int cmd_len = command_serialize(COMMAND_TYPE_RESPONSE, "counter_stream", 14,
                                    pb_buf, (uint16_t)ostream.bytes_written,
                                    cmd_buf, sizeof(cmd_buf));
    if (cmd_len < 0) {
        return -1;
    }

    /* Send via containers with retry */
    uint8_t tid = ble_service_next_transaction_id();
    uint16_t mtu = ble_service_get_mtu();
    return container_split_and_send(tid, cmd_buf, (size_t)cmd_len, mtu,
                                    notify_send_cb, NULL);
}

int handle_counter_stream(const uint8_t *req_data, size_t req_len, pb_ostream_t *ostream)
{
    (void)ostream; /* Not used — we send responses directly */

    blerpc_CounterStreamRequest req = blerpc_CounterStreamRequest_init_zero;
    pb_istream_t stream = pb_istream_from_buffer(req_data, req_len);

    if (!pb_decode(&stream, blerpc_CounterStreamRequest_fields, &req)) {
        LOG_ERR("CounterStream decode failed: %s", PB_GET_ERROR(&stream));
        return -1;
    }

    LOG_INF("CounterStream: count=%u", req.count);

    /* Send N responses, each with its own transaction_id */
    for (uint32_t i = 0; i < req.count; i++) {
        int rc = send_one_counter_stream_response(i, (int32_t)(i * 10));
        if (rc != 0) {
            LOG_ERR("CounterStream send %u failed: %d", i, rc);
            return -1;
        }
    }

    /* Send STREAM_END_P2C */
    uint8_t tid = ble_service_next_transaction_id();
    ble_service_send_stream_end_p2c(tid);

    /* Return -2: process_request will skip normal response */
    return -2;
}

/* ── counter_upload: C→P stream (accumulation) ────────────────────── */

static volatile uint32_t upload_count;

static void send_upload_response(struct k_work *work);
static K_WORK_DEFINE(upload_response_work, send_upload_response);

static void on_stream_end_c2p(uint8_t transaction_id)
{
    (void)transaction_id;
    LOG_INF("STREAM_END_C2P received, upload_count=%u", upload_count);
    ble_service_submit_work(&upload_response_work);
}

int handle_counter_upload(const uint8_t *req_data, size_t req_len, pb_ostream_t *ostream)
{
    (void)ostream;

    blerpc_CounterUploadRequest req = blerpc_CounterUploadRequest_init_zero;
    pb_istream_t stream = pb_istream_from_buffer(req_data, req_len);

    if (!pb_decode(&stream, blerpc_CounterUploadRequest_fields, &req)) {
        LOG_ERR("CounterUpload decode failed: %s", PB_GET_ERROR(&stream));
        return -1;
    }

    upload_count++;
    LOG_DBG("CounterUpload: seq=%u value=%d (total=%u)", req.seq, req.value, upload_count);

    /* Return -2: no response for individual stream messages */
    return -2;
}

static void send_upload_response(struct k_work *work)
{
    (void)work;

    uint32_t count = upload_count;
    upload_count = 0;

    LOG_INF("CounterUpload: sending response, received_count=%u", count);

    /* Encode CounterUploadResponse */
    blerpc_CounterUploadResponse resp = blerpc_CounterUploadResponse_init_zero;
    resp.received_count = count;

    uint8_t pb_buf[blerpc_CounterUploadResponse_size];
    pb_ostream_t ostream = pb_ostream_from_buffer(pb_buf, sizeof(pb_buf));
    if (!pb_encode(&ostream, blerpc_CounterUploadResponse_fields, &resp)) {
        LOG_ERR("CounterUploadResponse encode failed");
        return;
    }

    /* Build command response */
    static uint8_t cmd_buf[64];
    int cmd_len = command_serialize(COMMAND_TYPE_RESPONSE, "counter_upload", 14,
                                    pb_buf, (uint16_t)ostream.bytes_written,
                                    cmd_buf, sizeof(cmd_buf));
    if (cmd_len < 0) {
        LOG_ERR("Command serialize failed");
        return;
    }

    /* Send via containers */
    uint8_t tid = ble_service_next_transaction_id();
    uint16_t mtu = ble_service_get_mtu();
    container_split_and_send(tid, cmd_buf, (size_t)cmd_len, mtu,
                             notify_send_cb, NULL);
}

void handlers_stream_init(void)
{
    ble_service_set_stream_end_cb(on_stream_end_c2p);
}

int handle_data_write(const uint8_t *req_data, size_t req_len, pb_ostream_t *ostream)
{
    struct data_write_decode_ctx decode_ctx = {.total_bytes = 0};

    blerpc_DataWriteRequest req = blerpc_DataWriteRequest_init_zero;
    req.data.funcs.decode = data_write_decode_cb;
    req.data.arg = &decode_ctx;

    pb_istream_t stream = pb_istream_from_buffer(req_data, req_len);
    if (!pb_decode(&stream, blerpc_DataWriteRequest_fields, &req)) {
        LOG_ERR("DataWrite decode failed: %s", PB_GET_ERROR(&stream));
        return -1;
    }

    LOG_INF("DataWrite: received %u bytes", decode_ctx.total_bytes);

    blerpc_DataWriteResponse resp = blerpc_DataWriteResponse_init_zero;
    resp.length = decode_ctx.total_bytes;

    if (!pb_encode(ostream, blerpc_DataWriteResponse_fields, &resp)) {
        LOG_ERR("DataWrite encode failed: %s", PB_GET_ERROR(ostream));
        return -1;
    }

    return 0;
}
