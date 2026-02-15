#include <zephyr/ztest.h>
#include <string.h>
#include "container.h"

ZTEST_SUITE(container_tests, NULL, NULL, NULL, NULL, NULL);

ZTEST(container_tests, test_parse_first_container)
{
    /* Build a FIRST container: txn=1, seq=0, type=FIRST, total_len=5, payload_len=5, "hello" */
    uint8_t data[] = {
        0x01,       /* transaction_id */
        0x00,       /* sequence_number */
        0x00,       /* flags: type=0b00, control_cmd=0, reserved=0 */
        0x05, 0x00, /* total_length = 5 (LE) */
        0x05,       /* payload_len */
        'h', 'e', 'l', 'l', 'o'
    };

    struct container_header hdr;
    int rc = container_parse_header(data, sizeof(data), &hdr);
    zassert_equal(rc, 0, "parse should succeed");
    zassert_equal(hdr.transaction_id, 1, "");
    zassert_equal(hdr.sequence_number, 0, "");
    zassert_equal(hdr.type, CONTAINER_TYPE_FIRST, "");
    zassert_equal(hdr.total_length, 5, "");
    zassert_equal(hdr.payload_len, 5, "");
    zassert_mem_equal(hdr.payload, "hello", 5, "");
}

ZTEST(container_tests, test_parse_subsequent_container)
{
    /* type=SUBSEQUENT: flags = 0b01 << 6 = 0x40 */
    uint8_t data[] = {
        0x02,       /* transaction_id */
        0x01,       /* sequence_number */
        0x40,       /* flags: type=0b01 */
        0x03,       /* payload_len */
        'a', 'b', 'c'
    };

    struct container_header hdr;
    int rc = container_parse_header(data, sizeof(data), &hdr);
    zassert_equal(rc, 0, "");
    zassert_equal(hdr.type, CONTAINER_TYPE_SUBSEQUENT, "");
    zassert_equal(hdr.payload_len, 3, "");
    zassert_mem_equal(hdr.payload, "abc", 3, "");
}

ZTEST(container_tests, test_parse_control_container)
{
    /* type=CONTROL(0b11), control_cmd=TIMEOUT(0x1) => flags = (0b11 << 6) | (0x1 << 2) = 0xC4 */
    uint8_t data[] = {
        0x05,       /* transaction_id */
        0x00,       /* sequence_number */
        0xC4,       /* flags: type=CONTROL, cmd=TIMEOUT */
        0x02,       /* payload_len */
        0xC8, 0x00  /* timeout_ms = 200 (LE) */
    };

    struct container_header hdr;
    int rc = container_parse_header(data, sizeof(data), &hdr);
    zassert_equal(rc, 0, "");
    zassert_equal(hdr.type, CONTAINER_TYPE_CONTROL, "");
    zassert_equal(hdr.control_cmd, CONTROL_CMD_TIMEOUT, "");
    zassert_equal(hdr.payload_len, 2, "");
    uint16_t timeout = (uint16_t)hdr.payload[0] | ((uint16_t)hdr.payload[1] << 8);
    zassert_equal(timeout, 200, "");
}

ZTEST(container_tests, test_parse_too_short)
{
    uint8_t data[] = {0x00, 0x00};
    struct container_header hdr;
    int rc = container_parse_header(data, sizeof(data), &hdr);
    zassert_equal(rc, -1, "should fail on short data");
}

ZTEST(container_tests, test_serialize_first_roundtrip)
{
    struct container_header hdr = {
        .transaction_id = 10,
        .sequence_number = 0,
        .type = CONTAINER_TYPE_FIRST,
        .control_cmd = CONTROL_CMD_NONE,
        .total_length = 3,
        .payload_len = 3,
        .payload = (const uint8_t *)"abc",
    };

    uint8_t buf[64];
    int n = container_serialize(&hdr, buf, sizeof(buf));
    zassert_true(n > 0, "serialize should succeed");
    zassert_equal(n, CONTAINER_FIRST_HEADER_SIZE + 3, "");

    struct container_header parsed;
    int rc = container_parse_header(buf, (size_t)n, &parsed);
    zassert_equal(rc, 0, "");
    zassert_equal(parsed.transaction_id, 10, "");
    zassert_equal(parsed.total_length, 3, "");
    zassert_mem_equal(parsed.payload, "abc", 3, "");
}

ZTEST(container_tests, test_serialize_subsequent_roundtrip)
{
    struct container_header hdr = {
        .transaction_id = 10,
        .sequence_number = 1,
        .type = CONTAINER_TYPE_SUBSEQUENT,
        .control_cmd = CONTROL_CMD_NONE,
        .payload_len = 2,
        .payload = (const uint8_t *)"xy",
    };

    uint8_t buf[64];
    int n = container_serialize(&hdr, buf, sizeof(buf));
    zassert_true(n > 0, "");

    struct container_header parsed;
    int rc = container_parse_header(buf, (size_t)n, &parsed);
    zassert_equal(rc, 0, "");
    zassert_equal(parsed.type, CONTAINER_TYPE_SUBSEQUENT, "");
    zassert_mem_equal(parsed.payload, "xy", 2, "");
}

ZTEST(container_tests, test_assembler_single)
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

    int rc = container_assembler_feed(&a, &hdr);
    zassert_equal(rc, 1, "should be complete");
    zassert_mem_equal(a.buf, "hello", 5, "");
}

ZTEST(container_tests, test_assembler_multi)
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
    int rc = container_assembler_feed(&a, &first);
    zassert_equal(rc, 0, "need more");

    struct container_header second = {
        .transaction_id = 1,
        .sequence_number = 1,
        .type = CONTAINER_TYPE_SUBSEQUENT,
        .payload_len = 4,
        .payload = (const uint8_t *)"o wo",
    };
    rc = container_assembler_feed(&a, &second);
    zassert_equal(rc, 1, "should be complete");
    zassert_mem_equal(a.buf, "hello wo", 8, "");
}

ZTEST(container_tests, test_assembler_sequence_gap)
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
        .sequence_number = 2,  /* gap: expected 1 */
        .type = CONTAINER_TYPE_SUBSEQUENT,
        .payload_len = 3,
        .payload = (const uint8_t *)"def",
    };
    int rc = container_assembler_feed(&a, &bad);
    zassert_equal(rc, -1, "should fail on sequence gap");
    zassert_false(a.active, "assembler should be reset");
}

/* Test split_and_send */
static uint8_t send_buf[2048];
static size_t send_buf_offset;
static int send_count;

static int mock_send(const uint8_t *data, size_t len, void *ctx)
{
    (void)ctx;
    if (send_buf_offset + len > sizeof(send_buf)) {
        return -1;
    }
    memcpy(send_buf + send_buf_offset, data, len);
    send_buf_offset += len;
    send_count++;
    return 0;
}

ZTEST(container_tests, test_split_and_send_small)
{
    send_buf_offset = 0;
    send_count = 0;

    uint8_t payload[] = "hello";
    int rc = container_split_and_send(0, payload, 5, 247, mock_send, NULL);
    zassert_equal(rc, 0, "");
    zassert_equal(send_count, 1, "small payload should be 1 container");

    /* Parse back and verify */
    struct container_header hdr;
    rc = container_parse_header(send_buf, send_buf_offset, &hdr);
    zassert_equal(rc, 0, "");
    zassert_equal(hdr.type, CONTAINER_TYPE_FIRST, "");
    zassert_equal(hdr.total_length, 5, "");
    zassert_mem_equal(hdr.payload, "hello", 5, "");
}

ZTEST(container_tests, test_split_and_send_large)
{
    send_buf_offset = 0;
    send_count = 0;

    /* 100 bytes with MTU=27 => effective=24, first_max=18, sub_max=20 */
    uint8_t payload[100];
    memset(payload, 0xAB, 100);

    int rc = container_split_and_send(5, payload, 100, 27, mock_send, NULL);
    zassert_equal(rc, 0, "");
    zassert_true(send_count > 1, "should require multiple containers");

    /* Verify by assembling */
    struct container_assembler a;
    container_assembler_init(&a);

    size_t off = 0;
    int result = 0;
    while (off < send_buf_offset && result == 0) {
        struct container_header hdr;
        rc = container_parse_header(send_buf + off, send_buf_offset - off, &hdr);
        zassert_equal(rc, 0, "");

        size_t pkt_size;
        if (hdr.type == CONTAINER_TYPE_FIRST) {
            pkt_size = CONTAINER_FIRST_HEADER_SIZE + hdr.payload_len;
        } else {
            pkt_size = CONTAINER_SUBSEQUENT_HEADER_SIZE + hdr.payload_len;
        }

        result = container_assembler_feed(&a, &hdr);
        off += pkt_size;
    }

    zassert_equal(result, 1, "should complete assembly");
    zassert_mem_equal(a.buf, payload, 100, "");
}
