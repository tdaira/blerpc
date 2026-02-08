/**
 * Standalone C unit tests for container and command layers.
 * Builds on macOS/Linux with: gcc -o test_standalone test_standalone.c ../src/container.c
 * ../src/command.c -I../src
 */
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <assert.h>
#include <stdint.h>

#include "container.h"
#include "command.h"

static int tests_run = 0;
static int tests_passed = 0;

#define TEST(name)                                                                                 \
    do {                                                                                           \
        printf("  %s... ", #name);                                                                 \
        tests_run++;                                                                               \
        name();                                                                                    \
        tests_passed++;                                                                            \
        printf("PASSED\n");                                                                        \
    } while (0)

/* ===== Container tests ===== */

static void test_parse_first_container(void)
{
    uint8_t data[] = {0x01, 0x00, 0x00, /* txn=1, seq=0, flags=FIRST */
                      0x05, 0x00,       /* total_length=5 (LE) */
                      0x05,             /* payload_len */
                      'h',  'e',  'l',  'l', 'o'};
    struct container_header hdr;
    assert(container_parse_header(data, sizeof(data), &hdr) == 0);
    assert(hdr.transaction_id == 1);
    assert(hdr.sequence_number == 0);
    assert(hdr.type == CONTAINER_TYPE_FIRST);
    assert(hdr.total_length == 5);
    assert(hdr.payload_len == 5);
    assert(memcmp(hdr.payload, "hello", 5) == 0);
}

static void test_parse_subsequent_container(void)
{
    uint8_t data[] = {0x02, 0x01, 0x40, /* txn=2, seq=1, type=SUBSEQUENT */
                      0x03, 'a',  'b',  'c'};
    struct container_header hdr;
    assert(container_parse_header(data, sizeof(data), &hdr) == 0);
    assert(hdr.type == CONTAINER_TYPE_SUBSEQUENT);
    assert(hdr.payload_len == 3);
    assert(memcmp(hdr.payload, "abc", 3) == 0);
}

static void test_parse_control_container(void)
{
    /* type=CONTROL(0b11), control_cmd=TIMEOUT(0x1) => flags = 0xC4 */
    uint8_t data[] = {0x05, 0x00, 0xC4, 0x02, 0xC8, 0x00};
    struct container_header hdr;
    assert(container_parse_header(data, sizeof(data), &hdr) == 0);
    assert(hdr.type == CONTAINER_TYPE_CONTROL);
    assert(hdr.control_cmd == CONTROL_CMD_TIMEOUT);
    assert(hdr.payload_len == 2);
    uint16_t timeout = (uint16_t)hdr.payload[0] | ((uint16_t)hdr.payload[1] << 8);
    assert(timeout == 200);
}

static void test_parse_too_short(void)
{
    uint8_t data[] = {0x00, 0x00};
    struct container_header hdr;
    assert(container_parse_header(data, sizeof(data), &hdr) == -1);
}

static void test_serialize_first_roundtrip(void)
{
    uint8_t payload[] = "abc";
    struct container_header hdr = {
        .transaction_id = 10,
        .sequence_number = 0,
        .type = CONTAINER_TYPE_FIRST,
        .control_cmd = CONTROL_CMD_NONE,
        .total_length = 3,
        .payload_len = 3,
        .payload = payload,
    };
    uint8_t buf[64];
    int n = container_serialize(&hdr, buf, sizeof(buf));
    assert(n > 0);
    assert(n == CONTAINER_FIRST_HEADER_SIZE + 3);

    struct container_header parsed;
    assert(container_parse_header(buf, (size_t)n, &parsed) == 0);
    assert(parsed.transaction_id == 10);
    assert(parsed.total_length == 3);
    assert(memcmp(parsed.payload, "abc", 3) == 0);
}

static void test_serialize_subsequent_roundtrip(void)
{
    uint8_t payload[] = "xy";
    struct container_header hdr = {
        .transaction_id = 10,
        .sequence_number = 1,
        .type = CONTAINER_TYPE_SUBSEQUENT,
        .payload_len = 2,
        .payload = payload,
    };
    uint8_t buf[64];
    int n = container_serialize(&hdr, buf, sizeof(buf));
    assert(n > 0);

    struct container_header parsed;
    assert(container_parse_header(buf, (size_t)n, &parsed) == 0);
    assert(parsed.type == CONTAINER_TYPE_SUBSEQUENT);
    assert(memcmp(parsed.payload, "xy", 2) == 0);
}

static void test_assembler_single(void)
{
    struct container_assembler a;
    container_assembler_init(&a);

    struct container_header hdr = {
        .transaction_id = 0,
        .sequence_number = 0,
        .type = CONTAINER_TYPE_FIRST,
        .total_length = 5,
        .payload_len = 5,
        .payload = (const uint8_t *)"hello",
    };
    assert(container_assembler_feed(&a, &hdr) == 1);
    assert(memcmp(a.buf, "hello", 5) == 0);
}

static void test_assembler_multi(void)
{
    struct container_assembler a;
    container_assembler_init(&a);

    struct container_header first = {
        .transaction_id = 1,
        .sequence_number = 0,
        .type = CONTAINER_TYPE_FIRST,
        .total_length = 8,
        .payload_len = 4,
        .payload = (const uint8_t *)"hell",
    };
    assert(container_assembler_feed(&a, &first) == 0);

    struct container_header second = {
        .transaction_id = 1,
        .sequence_number = 1,
        .type = CONTAINER_TYPE_SUBSEQUENT,
        .payload_len = 4,
        .payload = (const uint8_t *)"o wo",
    };
    assert(container_assembler_feed(&a, &second) == 1);
    assert(memcmp(a.buf, "hello wo", 8) == 0);
}

static void test_assembler_sequence_gap(void)
{
    struct container_assembler a;
    container_assembler_init(&a);

    struct container_header first = {
        .transaction_id = 2,
        .sequence_number = 0,
        .type = CONTAINER_TYPE_FIRST,
        .total_length = 10,
        .payload_len = 3,
        .payload = (const uint8_t *)"abc",
    };
    container_assembler_feed(&a, &first);

    struct container_header bad = {
        .transaction_id = 2,
        .sequence_number = 2, /* gap: expected 1 */
        .type = CONTAINER_TYPE_SUBSEQUENT,
        .payload_len = 3,
        .payload = (const uint8_t *)"def",
    };
    assert(container_assembler_feed(&a, &bad) == -1);
    assert(a.active == false);
}

static uint8_t send_buf[2048];
static size_t send_buf_offset;
static int send_count;

static int mock_send(const uint8_t *data, size_t len, void *ctx)
{
    (void)ctx;
    if (send_buf_offset + len > sizeof(send_buf))
        return -1;
    memcpy(send_buf + send_buf_offset, data, len);
    send_buf_offset += len;
    send_count++;
    return 0;
}

static void test_split_and_send_small(void)
{
    send_buf_offset = 0;
    send_count = 0;
    uint8_t payload[] = "hello";
    assert(container_split_and_send(0, payload, 5, 247, mock_send, NULL) == 0);
    assert(send_count == 1);

    struct container_header hdr;
    assert(container_parse_header(send_buf, send_buf_offset, &hdr) == 0);
    assert(hdr.type == CONTAINER_TYPE_FIRST);
    assert(hdr.total_length == 5);
    assert(memcmp(hdr.payload, "hello", 5) == 0);
}

static void test_split_and_send_large(void)
{
    send_buf_offset = 0;
    send_count = 0;
    uint8_t payload[100];
    memset(payload, 0xAB, 100);

    assert(container_split_and_send(5, payload, 100, 27, mock_send, NULL) == 0);
    assert(send_count > 1);

    /* Verify by assembling */
    struct container_assembler a;
    container_assembler_init(&a);
    size_t off = 0;
    int result = 0;
    while (off < send_buf_offset && result == 0) {
        struct container_header hdr;
        assert(container_parse_header(send_buf + off, send_buf_offset - off, &hdr) == 0);
        size_t pkt_size = (hdr.type == CONTAINER_TYPE_FIRST)
                              ? CONTAINER_FIRST_HEADER_SIZE + hdr.payload_len
                              : CONTAINER_SUBSEQUENT_HEADER_SIZE + hdr.payload_len;
        result = container_assembler_feed(&a, &hdr);
        off += pkt_size;
    }
    assert(result == 1);
    assert(memcmp(a.buf, payload, 100) == 0);
}

/* ===== Command tests ===== */

static void test_command_parse_request(void)
{
    uint8_t data[] = {0x00, 0x04, 'e', 'c', 'h', 'o', 0x02, 0x00, 0x01, 0x02};
    struct command_packet pkt;
    assert(command_parse(data, sizeof(data), &pkt) == 0);
    assert(pkt.cmd_type == COMMAND_TYPE_REQUEST);
    assert(pkt.cmd_name_len == 4);
    assert(memcmp(pkt.cmd_name, "echo", 4) == 0);
    assert(pkt.data_len == 2);
    assert(pkt.data[0] == 0x01);
    assert(pkt.data[1] == 0x02);
}

static void test_command_parse_response(void)
{
    uint8_t data[] = {0x80, 0x04, 'e', 'c', 'h', 'o', 0x01, 0x00, 0xFF};
    struct command_packet pkt;
    assert(command_parse(data, sizeof(data), &pkt) == 0);
    assert(pkt.cmd_type == COMMAND_TYPE_RESPONSE);
    assert(pkt.data[0] == 0xFF);
}

static void test_command_serialize_roundtrip(void)
{
    uint8_t buf[128];
    uint8_t payload[] = {0xAA, 0xBB, 0xCC};
    int n = command_serialize(COMMAND_TYPE_REQUEST, "flash_read", 10, payload, 3, buf, sizeof(buf));
    assert(n > 0);

    struct command_packet pkt;
    assert(command_parse(buf, (size_t)n, &pkt) == 0);
    assert(pkt.cmd_type == COMMAND_TYPE_REQUEST);
    assert(pkt.cmd_name_len == 10);
    assert(memcmp(pkt.cmd_name, "flash_read", 10) == 0);
    assert(pkt.data_len == 3);
    assert(memcmp(pkt.data, payload, 3) == 0);
}

static void test_command_empty_data(void)
{
    uint8_t buf[64];
    int n = command_serialize(COMMAND_TYPE_REQUEST, "ping", 4, NULL, 0, buf, sizeof(buf));
    assert(n > 0);
    struct command_packet pkt;
    assert(command_parse(buf, (size_t)n, &pkt) == 0);
    assert(pkt.data_len == 0);
}

static void test_command_parse_too_short(void)
{
    uint8_t data[] = {0x00};
    struct command_packet pkt;
    assert(command_parse(data, sizeof(data), &pkt) == -1);
}

static void test_command_data_len_little_endian(void)
{
    uint8_t buf[512];
    uint8_t payload[300];
    memset(payload, 0, sizeof(payload));
    int n = command_serialize(COMMAND_TYPE_REQUEST, "x", 1, payload, 300, buf, sizeof(buf));
    assert(n > 0);
    /* data_len at offset 2+1=3, LE: 300 = 0x012C */
    assert(buf[3] == 0x2C);
    assert(buf[4] == 0x01);
}

int main(void)
{
    printf("Container tests:\n");
    TEST(test_parse_first_container);
    TEST(test_parse_subsequent_container);
    TEST(test_parse_control_container);
    TEST(test_parse_too_short);
    TEST(test_serialize_first_roundtrip);
    TEST(test_serialize_subsequent_roundtrip);
    TEST(test_assembler_single);
    TEST(test_assembler_multi);
    TEST(test_assembler_sequence_gap);
    TEST(test_split_and_send_small);
    TEST(test_split_and_send_large);

    printf("\nCommand tests:\n");
    TEST(test_command_parse_request);
    TEST(test_command_parse_response);
    TEST(test_command_serialize_roundtrip);
    TEST(test_command_empty_data);
    TEST(test_command_parse_too_short);
    TEST(test_command_data_len_little_endian);

    printf("\n%d/%d tests passed\n", tests_passed, tests_run);
    return (tests_passed == tests_run) ? 0 : 1;
}
