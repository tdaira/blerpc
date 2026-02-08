#ifndef BLERPC_COMMAND_H
#define BLERPC_COMMAND_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Command types */
#define COMMAND_TYPE_REQUEST 0
#define COMMAND_TYPE_RESPONSE 1

/**
 * Parsed command packet.
 * data pointer is zero-copy into source buffer.
 */
struct command_packet {
    uint8_t cmd_type; /* COMMAND_TYPE_* */
    uint8_t cmd_name_len;
    const char *cmd_name; /* Points into source buffer, NOT null-terminated */
    uint16_t data_len;
    const uint8_t *data; /* Points into source buffer */
};

/**
 * Parse a command packet from a buffer.
 * Zero-copy: cmd_name and data point into src.
 * Returns 0 on success, -1 on error.
 */
int command_parse(const uint8_t *src, size_t src_len, struct command_packet *out);

/**
 * Serialize a command packet into a buffer.
 * Returns number of bytes written, or -1 on error.
 */
int command_serialize(uint8_t cmd_type, const char *cmd_name, uint8_t cmd_name_len,
                      const uint8_t *data, uint16_t data_len, uint8_t *buf, size_t buf_size);

#ifdef __cplusplus
}
#endif

#endif /* BLERPC_COMMAND_H */
