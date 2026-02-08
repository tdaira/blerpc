#include "command.h"
#include <string.h>

int command_parse(const uint8_t *src, size_t src_len, struct command_packet *out)
{
    if (!src || !out || src_len < 2) {
        return -1;
    }

    /* Byte 0: type in MSB (bit 7) */
    out->cmd_type = (src[0] >> 7) & 0x01;
    out->cmd_name_len = src[1];

    size_t offset = 2;
    if (src_len < offset + out->cmd_name_len + 2) {
        return -1;
    }

    out->cmd_name = (const char *)(src + offset);
    offset += out->cmd_name_len;

    /* data_len: little-endian uint16 */
    out->data_len = (uint16_t)src[offset] | ((uint16_t)src[offset + 1] << 8);
    offset += 2;

    if (src_len < offset + out->data_len) {
        return -1;
    }

    out->data = src + offset;
    return 0;
}

int command_serialize(uint8_t cmd_type, const char *cmd_name, uint8_t cmd_name_len,
                      const uint8_t *data, uint16_t data_len, uint8_t *buf, size_t buf_size)
{
    size_t total = 2 + cmd_name_len + 2 + data_len;
    if (!buf || buf_size < total) {
        return -1;
    }

    /* Byte 0: type in MSB */
    buf[0] = (cmd_type & 0x01) << 7;
    buf[1] = cmd_name_len;

    if (cmd_name_len > 0 && cmd_name) {
        memcpy(buf + 2, cmd_name, cmd_name_len);
    }

    size_t offset = 2 + cmd_name_len;
    buf[offset] = (uint8_t)(data_len & 0xFF);
    buf[offset + 1] = (uint8_t)(data_len >> 8);
    offset += 2;

    if (data_len > 0 && data) {
        memcpy(buf + offset, data, data_len);
    }

    return (int)total;
}
