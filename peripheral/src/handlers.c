#include "handlers.h"
#include "blerpc.pb.h"
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
