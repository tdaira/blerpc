#include "container.h"
#include <string.h>

static uint8_t pack_flags(uint8_t type, uint8_t control_cmd)
{
    return ((type & 0x03) << 6) | ((control_cmd & 0x0F) << 2);
}

static void unpack_flags(uint8_t flags, uint8_t *type, uint8_t *control_cmd)
{
    *type = (flags >> 6) & 0x03;
    *control_cmd = (flags >> 2) & 0x0F;
}

int container_parse_header(const uint8_t *data, size_t len, struct container_header *out)
{
    if (!data || !out || len < 4) {
        return -1;
    }

    out->transaction_id = data[0];
    out->sequence_number = data[1];
    unpack_flags(data[2], &out->type, &out->control_cmd);

    if (out->type == CONTAINER_TYPE_FIRST) {
        if (len < CONTAINER_FIRST_HEADER_SIZE) {
            return -1;
        }
        out->total_length = (uint16_t)data[3] | ((uint16_t)data[4] << 8);
        out->payload_len = data[5];
        out->payload = data + CONTAINER_FIRST_HEADER_SIZE;

        if (len < CONTAINER_FIRST_HEADER_SIZE + out->payload_len) {
            return -1;
        }
    } else {
        out->total_length = 0;
        out->payload_len = data[3];
        out->payload = data + CONTAINER_SUBSEQUENT_HEADER_SIZE;

        if (len < CONTAINER_SUBSEQUENT_HEADER_SIZE + out->payload_len) {
            return -1;
        }
    }

    return 0;
}

int container_serialize(const struct container_header *hdr, uint8_t *buf, size_t buf_size)
{
    if (!hdr || !buf) {
        return -1;
    }

    size_t header_size;
    if (hdr->type == CONTAINER_TYPE_FIRST) {
        header_size = CONTAINER_FIRST_HEADER_SIZE;
    } else {
        header_size = CONTAINER_SUBSEQUENT_HEADER_SIZE;
    }

    size_t total = header_size + hdr->payload_len;
    if (buf_size < total) {
        return -1;
    }

    buf[0] = hdr->transaction_id;
    buf[1] = hdr->sequence_number;
    buf[2] = pack_flags(hdr->type, hdr->control_cmd);

    if (hdr->type == CONTAINER_TYPE_FIRST) {
        buf[3] = (uint8_t)(hdr->total_length & 0xFF);
        buf[4] = (uint8_t)(hdr->total_length >> 8);
        buf[5] = hdr->payload_len;
    } else {
        buf[3] = hdr->payload_len;
    }

    if (hdr->payload_len > 0 && hdr->payload) {
        memcpy(buf + header_size, hdr->payload, hdr->payload_len);
    }

    return (int)total;
}

void container_assembler_init(struct container_assembler *a)
{
    if (!a)
        return;
    memset(a, 0, sizeof(*a));
    a->active = false;
}

int container_assembler_feed(struct container_assembler *a, const struct container_header *hdr)
{
    if (!a || !hdr) {
        return -1;
    }

    /* Ignore control containers */
    if (hdr->type == CONTAINER_TYPE_CONTROL) {
        return 0;
    }

    if (hdr->type == CONTAINER_TYPE_FIRST) {
        a->transaction_id = hdr->transaction_id;
        a->total_length = hdr->total_length;
        a->expected_seq = 1;
        a->received_length = 0;
        a->active = true;

        if (hdr->payload_len > 0) {
            if ((size_t)hdr->payload_len > CONTAINER_ASSEMBLER_BUF_SIZE) {
                container_assembler_init(a);
                return -1;
            }
            memcpy(a->buf, hdr->payload, hdr->payload_len);
            a->received_length = hdr->payload_len;
        }
    } else if (hdr->type == CONTAINER_TYPE_SUBSEQUENT) {
        if (!a->active || hdr->transaction_id != a->transaction_id) {
            container_assembler_init(a);
            return -1;
        }
        if (hdr->sequence_number != a->expected_seq) {
            container_assembler_init(a);
            return -1;
        }
        if (a->received_length + hdr->payload_len > CONTAINER_ASSEMBLER_BUF_SIZE) {
            container_assembler_init(a);
            return -1;
        }
        memcpy(a->buf + a->received_length, hdr->payload, hdr->payload_len);
        a->received_length += hdr->payload_len;
        a->expected_seq++;
    }

    if (a->received_length >= a->total_length) {
        a->active = false;
        return 1; /* Assembly complete */
    }

    return 0; /* More containers needed */
}

int container_split_and_send(uint8_t transaction_id, const uint8_t *payload, size_t payload_len,
                             uint16_t mtu, container_send_fn send_fn, void *ctx)
{
    if (!send_fn) {
        return -1;
    }

    uint16_t effective_mtu = mtu - CONTAINER_ATT_OVERHEAD;
    uint8_t buf[256]; /* Max single BLE packet */

    /* First container */
    uint8_t first_max = (uint8_t)(effective_mtu - CONTAINER_FIRST_HEADER_SIZE);
    uint8_t first_len = (payload_len < first_max) ? (uint8_t)payload_len : first_max;

    struct container_header hdr = {
        .transaction_id = transaction_id,
        .sequence_number = 0,
        .type = CONTAINER_TYPE_FIRST,
        .control_cmd = CONTROL_CMD_NONE,
        .total_length = (uint16_t)payload_len,
        .payload_len = first_len,
        .payload = payload,
    };

    int n = container_serialize(&hdr, buf, sizeof(buf));
    if (n < 0)
        return -1;
    int rc = send_fn(buf, (size_t)n, ctx);
    if (rc < 0)
        return rc;

    size_t offset = first_len;
    uint8_t seq = 1;
    uint8_t sub_max = (uint8_t)(effective_mtu - CONTAINER_SUBSEQUENT_HEADER_SIZE);

    while (offset < payload_len) {
        uint8_t chunk_len =
            ((payload_len - offset) < sub_max) ? (uint8_t)(payload_len - offset) : sub_max;

        struct container_header sub_hdr = {
            .transaction_id = transaction_id,
            .sequence_number = seq,
            .type = CONTAINER_TYPE_SUBSEQUENT,
            .control_cmd = CONTROL_CMD_NONE,
            .total_length = 0,
            .payload_len = chunk_len,
            .payload = payload + offset,
        };

        n = container_serialize(&sub_hdr, buf, sizeof(buf));
        if (n < 0)
            return -1;
        rc = send_fn(buf, (size_t)n, ctx);
        if (rc < 0)
            return rc;

        offset += chunk_len;
        seq++;
    }

    return 0;
}
