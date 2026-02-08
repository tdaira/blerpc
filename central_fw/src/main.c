#include <zephyr/kernel.h>
#include <zephyr/bluetooth/bluetooth.h>
#include <zephyr/logging/log.h>
#include <pb_encode.h>
#include <pb_decode.h>
#include <string.h>

#include "ble_central.h"
#include <blerpc_protocol/container.h>
#include <blerpc_protocol/command.h>
#include "blerpc.pb.h"

LOG_MODULE_REGISTER(main, LOG_LEVEL_INF);

/* Response buffer and synchronization */
static uint8_t response_buf[12288];
static size_t response_len;
static K_SEM_DEFINE(response_sem, 0, 1);

static uint8_t transaction_counter;

static uint8_t next_transaction_id(void)
{
    return transaction_counter++;
}

/* Callback from ble_central when a complete response is assembled */
static void on_response(const uint8_t *data, size_t len)
{
    if (len > sizeof(response_buf)) {
        LOG_ERR("Response too large: %zu > %zu", len, sizeof(response_buf));
        return;
    }
    memcpy(response_buf, data, len);
    response_len = len;
    k_sem_give(&response_sem);
}

/* Send container callback for container_split_and_send */
static int send_container(const uint8_t *data, size_t len, void *ctx)
{
    (void)ctx;
    return ble_central_write(data, len);
}

/**
 * Send an RPC request and wait for the response.
 * Returns the protobuf response data length, or -1 on error.
 * Response protobuf data is in resp_data, length in resp_data_len.
 */
static int rpc_call(const char *cmd_name, const uint8_t *req_pb, size_t req_pb_len,
                    uint8_t *resp_pb, size_t resp_pb_size, size_t *resp_pb_len)
{
    static uint8_t cmd_buf[12288];
    uint8_t name_len = (uint8_t)strlen(cmd_name);

    int cmd_len = command_serialize(COMMAND_TYPE_REQUEST, cmd_name, name_len, req_pb,
                                    (uint16_t)req_pb_len, cmd_buf, sizeof(cmd_buf));
    if (cmd_len < 0) {
        LOG_ERR("Command serialize failed");
        return -1;
    }

    uint8_t tid = next_transaction_id();
    uint16_t mtu = ble_central_get_mtu();

    int rc = container_split_and_send(tid, cmd_buf, (size_t)cmd_len, mtu, send_container, NULL);
    if (rc < 0) {
        LOG_ERR("Container split/send failed: %d", rc);
        return -1;
    }

    /* Wait for response */
    rc = k_sem_take(&response_sem, K_SECONDS(10));
    if (rc != 0) {
        LOG_ERR("Response timeout");
        return -1;
    }

    /* Parse command response */
    struct command_packet resp_cmd;
    if (command_parse(response_buf, response_len, &resp_cmd) != 0) {
        LOG_ERR("Response command parse failed");
        return -1;
    }

    if (resp_cmd.cmd_type != COMMAND_TYPE_RESPONSE) {
        LOG_ERR("Expected response, got type %d", resp_cmd.cmd_type);
        return -1;
    }

    if (resp_cmd.cmd_name_len != name_len || memcmp(resp_cmd.cmd_name, cmd_name, name_len) != 0) {
        LOG_ERR("Command name mismatch in response");
        return -1;
    }

    if (resp_cmd.data_len > resp_pb_size) {
        LOG_ERR("Response data too large: %u > %zu", resp_cmd.data_len, resp_pb_size);
        return -1;
    }

    memcpy(resp_pb, resp_cmd.data, resp_cmd.data_len);
    *resp_pb_len = resp_cmd.data_len;
    return 0;
}

/* ── Test functions ──────────────────────────────────────────────────── */

static int test_echo(void)
{
    LOG_INF("=== Echo Test ===");

    const char *msg = "Hello from Central!";

    /* Encode request */
    static blerpc_EchoRequest req;
    memset(&req, 0, sizeof(req));
    strncpy(req.message, msg, sizeof(req.message) - 1);

    static uint8_t req_buf[blerpc_EchoRequest_size];
    pb_ostream_t ostream = pb_ostream_from_buffer(req_buf, sizeof(req_buf));
    if (!pb_encode(&ostream, blerpc_EchoRequest_fields, &req)) {
        LOG_ERR("Echo request encode failed");
        return -1;
    }

    /* RPC call */
    static uint8_t echo_resp_buf[blerpc_EchoResponse_size];
    size_t resp_len;
    if (rpc_call("echo", req_buf, ostream.bytes_written, echo_resp_buf, sizeof(echo_resp_buf),
                 &resp_len) != 0) {
        LOG_ERR("Echo RPC failed");
        return -1;
    }

    /* Decode response */
    static blerpc_EchoResponse resp;
    memset(&resp, 0, sizeof(resp));
    pb_istream_t istream = pb_istream_from_buffer(echo_resp_buf, resp_len);
    if (!pb_decode(&istream, blerpc_EchoResponse_fields, &resp)) {
        LOG_ERR("Echo response decode failed");
        return -1;
    }

    LOG_INF("Echo response: '%s'", resp.message);

    if (strcmp(resp.message, msg) != 0) {
        LOG_ERR("Echo mismatch! Expected '%s', got '%s'", msg, resp.message);
        return -1;
    }

    LOG_INF("Echo test PASSED");
    return 0;
}

/* Callback for decoding FlashReadResponse.data (FT_CALLBACK field) */
struct flash_read_decode_ctx {
    uint8_t *buf;
    size_t buf_size;
    size_t decoded_len;
};

static bool flash_data_decode_cb(pb_istream_t *stream, const pb_field_t *field, void **arg)
{
    (void)field;
    struct flash_read_decode_ctx *ctx = (struct flash_read_decode_ctx *)*arg;

    size_t len = stream->bytes_left;
    if (len > ctx->buf_size - ctx->decoded_len) {
        LOG_ERR("Flash data too large");
        return false;
    }

    if (!pb_read(stream, ctx->buf + ctx->decoded_len, len)) {
        return false;
    }
    ctx->decoded_len += len;
    return true;
}

static int test_flash_read(uint32_t length)
{
    LOG_INF("=== FlashRead Test (len=%u) ===", length);

    /* Encode request */
    blerpc_FlashReadRequest req = blerpc_FlashReadRequest_init_zero;
    req.address = 0x00000000;
    req.length = length;

    static uint8_t req_buf[blerpc_FlashReadRequest_size];
    pb_ostream_t ostream = pb_ostream_from_buffer(req_buf, sizeof(req_buf));
    if (!pb_encode(&ostream, blerpc_FlashReadRequest_fields, &req)) {
        LOG_ERR("FlashRead request encode failed");
        return -1;
    }

    /* RPC call */
    static uint8_t fr_resp_buf[10240];
    size_t resp_len;
    if (rpc_call("flash_read", req_buf, ostream.bytes_written, fr_resp_buf, sizeof(fr_resp_buf),
                 &resp_len) != 0) {
        LOG_ERR("FlashRead RPC failed");
        return -1;
    }

    /* Decode response */
    static uint8_t data_buf[8192];
    struct flash_read_decode_ctx decode_ctx = {
        .buf = data_buf,
        .buf_size = sizeof(data_buf),
        .decoded_len = 0,
    };

    blerpc_FlashReadResponse resp = blerpc_FlashReadResponse_init_zero;
    resp.data.funcs.decode = flash_data_decode_cb;
    resp.data.arg = &decode_ctx;

    pb_istream_t istream = pb_istream_from_buffer(fr_resp_buf, resp_len);
    if (!pb_decode(&istream, blerpc_FlashReadResponse_fields, &resp)) {
        LOG_ERR("FlashRead response decode failed");
        return -1;
    }

    LOG_INF("FlashRead response: addr=0x%08x, data_len=%zu", resp.address, decode_ctx.decoded_len);

    if (decode_ctx.decoded_len != length) {
        LOG_ERR("FlashRead length mismatch: expected %u, got %zu", length, decode_ctx.decoded_len);
        return -1;
    }

    LOG_INF("FlashRead test PASSED");
    return 0;
}

static int test_throughput(void)
{
    LOG_INF("=== Throughput Test (10x flash_read 8192) ===");

    /* Warm up */
    if (test_flash_read(8192) != 0) {
        LOG_ERR("Throughput warm-up failed");
        return -1;
    }

    uint32_t start = k_uptime_get_32();

    for (int i = 0; i < 10; i++) {
        if (test_flash_read(8192) != 0) {
            LOG_ERR("Throughput test failed at iteration %d", i);
            return -1;
        }
    }

    uint32_t elapsed = k_uptime_get_32() - start;
    uint32_t total_bytes = 10 * 8192;
    uint32_t kbps = (total_bytes * 1000) / (elapsed * 1024);

    LOG_INF("Throughput: %u bytes in %u ms = %u KB/s", total_bytes, elapsed, kbps);
    LOG_INF("Throughput test PASSED");
    return 0;
}

/* Callback for encoding DataWriteRequest.data (FT_CALLBACK field) */
struct data_write_encode_ctx {
    uint32_t length;
};

static bool data_write_encode_cb(pb_ostream_t *stream, const pb_field_t *field, void *const *arg)
{
    struct data_write_encode_ctx *ctx = *(struct data_write_encode_ctx **)arg;

    if (!pb_encode_tag_for_field(stream, field))
        return false;
    if (!pb_encode_varint(stream, ctx->length))
        return false;

    /* Write incrementing pattern bytes */
    uint8_t chunk[256];
    uint32_t remaining = ctx->length;
    uint32_t offset = 0;
    while (remaining > 0) {
        uint32_t n = MIN(remaining, sizeof(chunk));
        for (uint32_t i = 0; i < n; i++) {
            chunk[i] = (uint8_t)((offset + i) & 0xFF);
        }
        if (!pb_write(stream, chunk, n))
            return false;
        offset += n;
        remaining -= n;
    }
    return true;
}

static int test_data_write(uint32_t length)
{
    LOG_INF("=== DataWrite Test (len=%u) ===", length);

    struct data_write_encode_ctx encode_ctx = {.length = length};

    blerpc_DataWriteRequest req = blerpc_DataWriteRequest_init_zero;
    req.data.funcs.encode = data_write_encode_cb;
    req.data.arg = &encode_ctx;

    /* Pass 1: sizing */
    pb_ostream_t sizing = PB_OSTREAM_SIZING;
    if (!pb_encode(&sizing, blerpc_DataWriteRequest_fields, &req)) {
        LOG_ERR("DataWrite request sizing failed");
        return -1;
    }

    /* Pass 2: encode */
    static uint8_t req_buf[10240];
    if (sizing.bytes_written > sizeof(req_buf)) {
        LOG_ERR("DataWrite request too large: %zu > %zu", sizing.bytes_written, sizeof(req_buf));
        return -1;
    }

    pb_ostream_t ostream = pb_ostream_from_buffer(req_buf, sizeof(req_buf));
    if (!pb_encode(&ostream, blerpc_DataWriteRequest_fields, &req)) {
        LOG_ERR("DataWrite request encode failed");
        return -1;
    }

    /* RPC call */
    static uint8_t dw_resp_buf[blerpc_DataWriteResponse_size];
    size_t resp_len;
    if (rpc_call("data_write", req_buf, ostream.bytes_written, dw_resp_buf, sizeof(dw_resp_buf),
                 &resp_len) != 0) {
        LOG_ERR("DataWrite RPC failed");
        return -1;
    }

    /* Decode response */
    blerpc_DataWriteResponse resp = blerpc_DataWriteResponse_init_zero;
    pb_istream_t istream = pb_istream_from_buffer(dw_resp_buf, resp_len);
    if (!pb_decode(&istream, blerpc_DataWriteResponse_fields, &resp)) {
        LOG_ERR("DataWrite response decode failed");
        return -1;
    }

    LOG_INF("DataWrite response: length=%u", resp.length);

    if (resp.length != length) {
        LOG_ERR("DataWrite length mismatch: expected %u, got %u", length, resp.length);
        return -1;
    }

    LOG_INF("DataWrite test PASSED");
    return 0;
}

static int test_write_throughput(void)
{
    LOG_INF("=== Write Throughput Test (10x data_write 8192) ===");

    /* Warm up */
    if (test_data_write(8192) != 0) {
        LOG_ERR("Write throughput warm-up failed");
        return -1;
    }

    uint32_t start = k_uptime_get_32();

    for (int i = 0; i < 10; i++) {
        if (test_data_write(8192) != 0) {
            LOG_ERR("Write throughput test failed at iteration %d", i);
            return -1;
        }
    }

    uint32_t elapsed = k_uptime_get_32() - start;
    uint32_t total_bytes = 10 * 8192;
    uint32_t kbps = (total_bytes * 1000) / (elapsed * 1024);

    LOG_INF("Write throughput: %u bytes in %u ms = %u KB/s", total_bytes, elapsed, kbps);
    LOG_INF("Write throughput test PASSED");
    return 0;
}

/* ── Main ────────────────────────────────────────────────────────────── */

int main(void)
{
    int err;

    LOG_INF("blerpc central starting");

    err = bt_enable(NULL);
    if (err) {
        LOG_ERR("Bluetooth init failed (err %d)", err);
        return err;
    }
    LOG_INF("Bluetooth initialized");

    ble_central_init(on_response);

    err = ble_central_connect();
    if (err) {
        LOG_ERR("Connect failed (err %d)", err);
        return err;
    }

    LOG_INF("MTU: %u", ble_central_get_mtu());

    /* Allow subscription to settle */
    k_sleep(K_MSEC(200));

    /* Run tests */
    int failures = 0;

    if (test_echo() != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_flash_read(8192) != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_throughput() != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_data_write(8192) != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_write_throughput() != 0) {
        failures++;
    }

    LOG_INF("===========================");
    if (failures == 0) {
        LOG_INF("All tests PASSED");
    } else {
        LOG_ERR("%d test(s) FAILED", failures);
    }

    return 0;
}
