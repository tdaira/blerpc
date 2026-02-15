#include <zephyr/ztest.h>
#include <string.h>
#include "command.h"

ZTEST_SUITE(command_tests, NULL, NULL, NULL, NULL, NULL);

ZTEST(command_tests, test_parse_request)
{
    /* type=REQUEST(0), cmd_name_len=4, "echo", data_len=2(LE), data=0x01,0x02 */
    uint8_t data[] = {
        0x00,                   /* byte 0: type=0 */
        0x04,                   /* cmd_name_len */
        'e', 'c', 'h', 'o',    /* cmd_name */
        0x02, 0x00,             /* data_len = 2 (LE) */
        0x01, 0x02              /* data */
    };

    struct command_packet pkt;
    int rc = command_parse(data, sizeof(data), &pkt);
    zassert_equal(rc, 0, "parse should succeed");
    zassert_equal(pkt.cmd_type, COMMAND_TYPE_REQUEST, "");
    zassert_equal(pkt.cmd_name_len, 4, "");
    zassert_mem_equal(pkt.cmd_name, "echo", 4, "");
    zassert_equal(pkt.data_len, 2, "");
    zassert_equal(pkt.data[0], 0x01, "");
    zassert_equal(pkt.data[1], 0x02, "");
}

ZTEST(command_tests, test_parse_response)
{
    /* type=RESPONSE(1) => byte 0 = 0x80 */
    uint8_t data[] = {
        0x80,                   /* byte 0: type=1 */
        0x04,
        'e', 'c', 'h', 'o',
        0x01, 0x00,
        0xFF
    };

    struct command_packet pkt;
    int rc = command_parse(data, sizeof(data), &pkt);
    zassert_equal(rc, 0, "");
    zassert_equal(pkt.cmd_type, COMMAND_TYPE_RESPONSE, "");
    zassert_equal(pkt.data[0], 0xFF, "");
}

ZTEST(command_tests, test_serialize_roundtrip)
{
    uint8_t buf[128];
    uint8_t payload[] = {0xAA, 0xBB, 0xCC};

    int n = command_serialize(COMMAND_TYPE_REQUEST, "flash_read", 10,
                              payload, 3, buf, sizeof(buf));
    zassert_true(n > 0, "serialize should succeed");

    struct command_packet pkt;
    int rc = command_parse(buf, (size_t)n, &pkt);
    zassert_equal(rc, 0, "");
    zassert_equal(pkt.cmd_type, COMMAND_TYPE_REQUEST, "");
    zassert_equal(pkt.cmd_name_len, 10, "");
    zassert_mem_equal(pkt.cmd_name, "flash_read", 10, "");
    zassert_equal(pkt.data_len, 3, "");
    zassert_mem_equal(pkt.data, payload, 3, "");
}

ZTEST(command_tests, test_serialize_response)
{
    uint8_t buf[64];
    int n = command_serialize(COMMAND_TYPE_RESPONSE, "echo", 4,
                              (const uint8_t *)"hi", 2, buf, sizeof(buf));
    zassert_true(n > 0, "");
    zassert_equal(buf[0], 0x80, "response type bit should be set");
}

ZTEST(command_tests, test_empty_data)
{
    uint8_t buf[64];
    int n = command_serialize(COMMAND_TYPE_REQUEST, "ping", 4,
                              NULL, 0, buf, sizeof(buf));
    zassert_true(n > 0, "");

    struct command_packet pkt;
    int rc = command_parse(buf, (size_t)n, &pkt);
    zassert_equal(rc, 0, "");
    zassert_equal(pkt.data_len, 0, "");
}

ZTEST(command_tests, test_parse_too_short)
{
    uint8_t data[] = {0x00};
    struct command_packet pkt;
    int rc = command_parse(data, sizeof(data), &pkt);
    zassert_equal(rc, -1, "should fail on short data");
}

ZTEST(command_tests, test_data_len_little_endian)
{
    uint8_t buf[512];
    uint8_t payload[300];
    memset(payload, 0, sizeof(payload));

    int n = command_serialize(COMMAND_TYPE_REQUEST, "x", 1,
                              payload, 300, buf, sizeof(buf));
    zassert_true(n > 0, "");

    /* data_len at offset 2 + 1 = 3, should be 300 = 0x012C in LE => 0x2C, 0x01 */
    zassert_equal(buf[3], 0x2C, "low byte");
    zassert_equal(buf[4], 0x01, "high byte");
}
