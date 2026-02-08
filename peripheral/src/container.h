#ifndef BLERPC_CONTAINER_H
#define BLERPC_CONTAINER_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Container types */
#define CONTAINER_TYPE_FIRST 0x00
#define CONTAINER_TYPE_SUBSEQUENT 0x01
#define CONTAINER_TYPE_CONTROL 0x03

/* Control commands */
#define CONTROL_CMD_NONE 0x00
#define CONTROL_CMD_TIMEOUT 0x01
#define CONTROL_CMD_STREAM_END_C2P 0x02
#define CONTROL_CMD_STREAM_END_P2C 0x03

/* Header sizes */
#define CONTAINER_FIRST_HEADER_SIZE 6
#define CONTAINER_SUBSEQUENT_HEADER_SIZE 4
#define CONTAINER_CONTROL_HEADER_SIZE 4
#define CONTAINER_ATT_OVERHEAD 3

/* Assembler buffer size (configurable) */
#ifndef CONTAINER_ASSEMBLER_BUF_SIZE
#define CONTAINER_ASSEMBLER_BUF_SIZE 4096
#endif

/**
 * Parsed container header.
 */
struct container_header {
    uint8_t transaction_id;
    uint8_t sequence_number;
    uint8_t type;          /* CONTAINER_TYPE_* */
    uint8_t control_cmd;   /* CONTROL_CMD_* */
    uint16_t total_length; /* Only valid for FIRST */
    uint8_t payload_len;
    const uint8_t *payload; /* Points into source buffer */
};

/**
 * Parse a raw BLE packet into a container header.
 * Returns 0 on success, -1 on error.
 * payload pointer will point into the source data buffer (zero-copy).
 */
int container_parse_header(const uint8_t *data, size_t len, struct container_header *out);

/**
 * Serialize a container header + payload into a buffer.
 * Returns number of bytes written, or -1 on error.
 */
int container_serialize(const struct container_header *hdr, uint8_t *buf, size_t buf_size);

/**
 * Container assembler state.
 */
struct container_assembler {
    uint8_t buf[CONTAINER_ASSEMBLER_BUF_SIZE];
    uint16_t total_length;
    uint16_t received_length;
    uint8_t expected_seq;
    uint8_t transaction_id;
    bool active;
};

/**
 * Initialize the assembler.
 */
void container_assembler_init(struct container_assembler *a);

/**
 * Feed a parsed container into the assembler.
 * Returns:
 *   0  if more containers needed
 *   1  if assembly complete (data in a->buf, length in a->total_length)
 *  -1  on error (sequence gap, overflow, etc.) — assembler is reset
 */
int container_assembler_feed(struct container_assembler *a, const struct container_header *hdr);

/**
 * Callback type for sending a container.
 * data: serialized container bytes, len: number of bytes.
 * Returns 0 on success, negative on error.
 */
typedef int (*container_send_fn)(const uint8_t *data, size_t len, void *ctx);

/**
 * Split payload into containers and send each via callback.
 * Returns 0 on success, negative on error.
 */
int container_split_and_send(uint8_t transaction_id, const uint8_t *payload, size_t payload_len,
                             uint16_t mtu, container_send_fn send_fn, void *ctx);

#ifdef __cplusplus
}
#endif

#endif /* BLERPC_CONTAINER_H */
