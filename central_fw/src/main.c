#include <zephyr/kernel.h>
#include <zephyr/bluetooth/bluetooth.h>
#include <zephyr/logging/log.h>
#include <string.h>

#include "ble_central.h"
#include <blerpc_protocol/container.h>
#include <blerpc_protocol/command.h>
#include "generated_client.h"
#ifdef CONFIG_BLERPC_ENCRYPTION
#include <mbedtls/platform.h>
#endif

LOG_MODULE_REGISTER(main, LOG_LEVEL_INF);

/* Max test payload adapts to assembler buffer size (leave headroom for headers) */
#define MAX_TEST_PAYLOAD (CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE - 128)

/* Response buffer and synchronization */
static uint8_t response_buf[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE];
static size_t response_len;

/* Shared encryption buffer (reused across sequential calls) */
static uint8_t encrypt_buf[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE + 20];

/* Shared work buffers (tests run sequentially, not concurrently) */
static uint8_t shared_cmd_buf[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE];
static uint8_t shared_work_buf[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE];
static uint8_t shared_decode_buf[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE];

static int rpc_error_code;
static K_SEM_DEFINE(response_sem, 0, 10);

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
    rpc_error_code = 0;
    k_sem_give(&response_sem);
}

/* Callback from ble_central when an ERROR control container is received */
static void on_error(uint8_t error_code)
{
    LOG_ERR("Peripheral error: 0x%02x", error_code);
    rpc_error_code = error_code;
    k_sem_give(&response_sem);
}

/* Send container callback for container_split_and_send */
static int send_container(const uint8_t *data, size_t len, void *ctx)
{
    (void)ctx;
    return ble_central_write(data, len);
}

/* ── RPC transport functions (extern'd by generated_client.h) ────────── */

int blerpc_rpc_call(const char *cmd_name, const uint8_t *req_data, size_t req_len,
                    uint8_t *resp_data, size_t resp_size, size_t *resp_len)
{
    uint8_t name_len = (uint8_t)strlen(cmd_name);

    int cmd_len = command_serialize(COMMAND_TYPE_REQUEST, cmd_name, name_len, req_data,
                                    (uint16_t)req_len, shared_cmd_buf, sizeof(shared_cmd_buf));
    if (cmd_len < 0) {
        LOG_ERR("Command serialize failed");
        return -1;
    }

    uint16_t max_req = ble_central_get_max_request_payload_size();
    if (max_req > 0 && (size_t)cmd_len > max_req) {
        LOG_ERR("Request too large: %d > %u", cmd_len, max_req);
        return -1;
    }

    /* Encrypt if encryption is active */
    size_t send_len;
    if (ble_central_encrypt_payload(shared_cmd_buf, (size_t)cmd_len, encrypt_buf,
                                    sizeof(encrypt_buf), &send_len) != 0) {
        LOG_ERR("Payload encryption failed");
        return -1;
    }

    uint8_t tid = next_transaction_id();
    uint16_t mtu = ble_central_get_mtu();

    int rc = container_split_and_send(tid, encrypt_buf, send_len, mtu, send_container, NULL);
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

    if (rpc_error_code != 0) {
        LOG_ERR("RPC error from peripheral: 0x%02x", rpc_error_code);
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

    if (resp_cmd.data_len > resp_size) {
        LOG_ERR("Response data too large: %u > %zu", resp_cmd.data_len, resp_size);
        return -1;
    }

    memcpy(resp_data, resp_cmd.data, resp_cmd.data_len);
    *resp_len = resp_cmd.data_len;
    return 0;
}

/* Stream end signaling for blerpc_stream_receive */
static volatile bool _stream_ended;

static void _stream_end_cb(void)
{
    _stream_ended = true;
    k_sem_give(&response_sem);
}

int blerpc_stream_receive(const char *cmd_name, const uint8_t *req_data, size_t req_len,
                          blerpc_on_stream_resp_t on_resp, void *ctx)
{
    uint8_t name_len = (uint8_t)strlen(cmd_name);

    _stream_ended = false;
    ble_central_set_stream_end_cb(_stream_end_cb);

    /* Serialize and send the initial request */
    int cmd_len = command_serialize(COMMAND_TYPE_REQUEST, cmd_name, name_len, req_data,
                                    (uint16_t)req_len, shared_cmd_buf, sizeof(shared_cmd_buf));
    if (cmd_len < 0) {
        LOG_ERR("Command serialize failed");
        ble_central_set_stream_end_cb(NULL);
        return -1;
    }

    size_t send_len;
    if (ble_central_encrypt_payload(shared_cmd_buf, (size_t)cmd_len, encrypt_buf,
                                    sizeof(encrypt_buf), &send_len) != 0) {
        LOG_ERR("Payload encryption failed");
        ble_central_set_stream_end_cb(NULL);
        return -1;
    }

    uint8_t tid = next_transaction_id();
    uint16_t mtu = ble_central_get_mtu();
    int rc = container_split_and_send(tid, encrypt_buf, send_len, mtu, send_container, NULL);
    if (rc < 0) {
        LOG_ERR("Container split/send failed: %d", rc);
        ble_central_set_stream_end_cb(NULL);
        return -1;
    }

    /* Receive responses until STREAM_END_P2C */
    while (true) {
        rc = k_sem_take(&response_sem, K_SECONDS(10));
        if (rc != 0) {
            LOG_ERR("Stream response timeout");
            ble_central_set_stream_end_cb(NULL);
            return -1;
        }

        if (_stream_ended) {
            break;
        }

        if (rpc_error_code != 0) {
            LOG_ERR("Stream error: 0x%02x", rpc_error_code);
            ble_central_set_stream_end_cb(NULL);
            return -1;
        }

        struct command_packet resp_cmd;
        if (command_parse(response_buf, response_len, &resp_cmd) != 0) {
            LOG_ERR("Stream response parse failed");
            ble_central_set_stream_end_cb(NULL);
            return -1;
        }

        if (on_resp(resp_cmd.data, resp_cmd.data_len, ctx) != 0) {
            LOG_ERR("Stream response callback failed");
            ble_central_set_stream_end_cb(NULL);
            return -1;
        }
    }

    ble_central_set_stream_end_cb(NULL);
    return 0;
}

int blerpc_stream_send(const char *cmd_name, size_t msg_count,
                       blerpc_next_msg_t next_msg, void *msg_ctx,
                       const char *final_cmd_name,
                       uint8_t *resp_data, size_t resp_size, size_t *resp_len)
{
    (void)final_cmd_name;
    uint8_t name_len = (uint8_t)strlen(cmd_name);

    /* Send each message */
    for (size_t i = 0; i < msg_count; i++) {
        size_t msg_len;
        if (next_msg(i, shared_work_buf, sizeof(shared_work_buf), &msg_len, msg_ctx) != 0) {
            LOG_ERR("next_msg callback failed at %zu", i);
            return -1;
        }

        int cmd_len = command_serialize(COMMAND_TYPE_REQUEST, cmd_name, name_len,
                                        shared_work_buf, (uint16_t)msg_len,
                                        shared_cmd_buf, sizeof(shared_cmd_buf));
        if (cmd_len < 0) {
            LOG_ERR("Command serialize failed at %zu", i);
            return -1;
        }

        size_t send_len;
        if (ble_central_encrypt_payload(shared_cmd_buf, (size_t)cmd_len, encrypt_buf,
                                        sizeof(encrypt_buf), &send_len) != 0) {
            LOG_ERR("Payload encryption failed at %zu", i);
            return -1;
        }

        uint8_t tid = next_transaction_id();
        uint16_t mtu = ble_central_get_mtu();
        int rc = container_split_and_send(tid, encrypt_buf, send_len, mtu, send_container, NULL);
        if (rc < 0) {
            LOG_ERR("Container split/send failed at %zu: %d", i, rc);
            return -1;
        }
    }

    /* Send STREAM_END_C2P */
    if (ble_central_send_stream_end_c2p() < 0) {
        LOG_ERR("STREAM_END_C2P send failed");
        return -1;
    }

    /* Wait for final response */
    int rc = k_sem_take(&response_sem, K_SECONDS(10));
    if (rc != 0) {
        LOG_ERR("Final response timeout");
        return -1;
    }

    if (rpc_error_code != 0) {
        LOG_ERR("RPC error from peripheral: 0x%02x", rpc_error_code);
        return -1;
    }

    struct command_packet resp_cmd;
    if (command_parse(response_buf, response_len, &resp_cmd) != 0) {
        LOG_ERR("Response command parse failed");
        return -1;
    }

    if (resp_cmd.data_len > resp_size) {
        LOG_ERR("Response data too large: %u > %zu", resp_cmd.data_len, resp_size);
        return -1;
    }

    memcpy(resp_data, resp_cmd.data, resp_cmd.data_len);
    *resp_len = resp_cmd.data_len;
    return 0;
}

/* ── Test functions ──────────────────────────────────────────────────── */

static int test_echo(void)
{
    LOG_INF("=== Echo Test ===");

    static const char msg[] = "Hello from nRF54L15 central!";

    blerpc_EchoResponse resp;
    if (blerpc_echo(msg, &resp) != 0) {
        LOG_ERR("Echo RPC failed");
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

static int test_flash_read(uint32_t length)
{
    LOG_INF("=== FlashRead Test (len=%u) ===", length);

    blerpc_FlashReadResponse resp;
    size_t data_len;
    if (blerpc_flash_read(0x00000000, length, &resp,
                          shared_decode_buf, sizeof(shared_decode_buf), &data_len) != 0) {
        LOG_ERR("FlashRead RPC failed");
        return -1;
    }

    LOG_INF("FlashRead response: addr=0x%08x, data_len=%zu", resp.address, data_len);

    if (data_len != length) {
        LOG_ERR("FlashRead length mismatch: expected %u, got %zu", length, data_len);
        return -1;
    }

    LOG_INF("FlashRead test PASSED");
    return 0;
}

static int test_throughput(void)
{
    LOG_INF("=== Throughput Test (10x flash_read %u) ===", MAX_TEST_PAYLOAD);

    /* Warm up */
    if (test_flash_read(MAX_TEST_PAYLOAD) != 0) {
        LOG_ERR("Throughput warm-up failed");
        return -1;
    }

    uint32_t start = k_uptime_get_32();

    for (int i = 0; i < 10; i++) {
        if (test_flash_read(MAX_TEST_PAYLOAD) != 0) {
            LOG_ERR("Throughput test failed at iteration %d", i);
            return -1;
        }
    }

    uint32_t elapsed = k_uptime_get_32() - start;
    uint32_t total_bytes = 10 * MAX_TEST_PAYLOAD;
    uint32_t kbps = (total_bytes * 1000) / (elapsed * 1024);

    LOG_INF("Throughput: %u bytes in %u ms = %u KB/s", total_bytes, elapsed, kbps);
    LOG_INF("Throughput test PASSED");
    return 0;
}

static int test_data_write(uint32_t length)
{
    LOG_INF("=== DataWrite Test (len=%u) ===", length);

    /* Build incrementing pattern data */
    for (uint32_t i = 0; i < length; i++) {
        shared_decode_buf[i] = (uint8_t)(i & 0xFF);
    }

    blerpc_DataWriteResponse resp;
    if (blerpc_data_write(shared_decode_buf, length,
                          shared_work_buf, sizeof(shared_work_buf), &resp) != 0) {
        LOG_ERR("DataWrite RPC failed");
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
    LOG_INF("=== Write Throughput Test (10x data_write %u) ===", MAX_TEST_PAYLOAD);

    /* Warm up */
    if (test_data_write(MAX_TEST_PAYLOAD) != 0) {
        LOG_ERR("Write throughput warm-up failed");
        return -1;
    }

    uint32_t start = k_uptime_get_32();

    for (int i = 0; i < 10; i++) {
        if (test_data_write(MAX_TEST_PAYLOAD) != 0) {
            LOG_ERR("Write throughput test failed at iteration %d", i);
            return -1;
        }
    }

    uint32_t elapsed = k_uptime_get_32() - start;
    uint32_t total_bytes = 10 * MAX_TEST_PAYLOAD;
    uint32_t kbps = (total_bytes * 1000) / (elapsed * 1024);

    LOG_INF("Write throughput: %u bytes in %u ms = %u KB/s", total_bytes, elapsed, kbps);
    LOG_INF("Write throughput test PASSED");
    return 0;
}

static int test_counter_stream(void)
{
    LOG_INF("=== CounterStream Test ===");

    const uint32_t count = 5;
    blerpc_CounterStreamResponse results[10];
    size_t result_count;

    if (blerpc_counter_stream(count, results, 10, &result_count) != 0) {
        LOG_ERR("CounterStream failed");
        return -1;
    }

    for (size_t i = 0; i < result_count; i++) {
        if (results[i].seq != i || results[i].value != (int32_t)(i * 10)) {
            LOG_ERR("CounterStream mismatch at %zu: seq=%u value=%d",
                    i, results[i].seq, results[i].value);
            return -1;
        }
    }

    LOG_INF("CounterStream: received %zu responses", result_count);
    if (result_count != count) {
        LOG_ERR("CounterStream count mismatch: expected %u, got %zu", count, result_count);
        return -1;
    }

    LOG_INF("CounterStream test PASSED");
    return 0;
}

static int test_counter_upload(void)
{
    LOG_INF("=== CounterUpload Test ===");

    const uint32_t count = 5;
    blerpc_CounterUploadRequest messages[5];
    for (uint32_t i = 0; i < count; i++) {
        messages[i] = (blerpc_CounterUploadRequest)blerpc_CounterUploadRequest_init_zero;
        messages[i].seq = i;
        messages[i].value = (int32_t)(i * 10);
    }

    blerpc_CounterUploadResponse resp;
    if (blerpc_counter_upload(messages, count, &resp) != 0) {
        LOG_ERR("CounterUpload failed");
        return -1;
    }

    LOG_INF("CounterUpload response: received_count=%u", resp.received_count);

    if (resp.received_count != count) {
        LOG_ERR("CounterUpload count mismatch: expected %u, got %u", count, resp.received_count);
        return -1;
    }

    LOG_INF("CounterUpload test PASSED");
    return 0;
}

/* ── Main ────────────────────────────────────────────────────────────── */

int main(void)
{
    int err;

    LOG_INF("blerpc central starting");

#ifdef CONFIG_BLERPC_ENCRYPTION
    /* mbedTLS on NCS defaults mbedtls_calloc to a stub returning NULL.
     * Use Zephyr's k_calloc/k_free backed by CONFIG_HEAP_MEM_POOL_SIZE. */
    mbedtls_platform_set_calloc_free(k_calloc, k_free);
#endif

    err = bt_enable(NULL);
    if (err) {
        LOG_ERR("Bluetooth init failed (err %d)", err);
        return err;
    }
    LOG_INF("Bluetooth initialized");

    ble_central_init(on_response, on_error);

    err = ble_central_connect();
    if (err) {
        LOG_ERR("Connect failed (err %d)", err);
        return err;
    }

    LOG_INF("MTU: %u", ble_central_get_mtu());

    /* Request capabilities from peripheral */
    err = ble_central_request_capabilities();
    if (err) {
        LOG_WRN("Capabilities request failed (err %d), continuing without limits", err);
    } else {
        LOG_INF("Peripheral capabilities: max_request=%u, max_response=%u",
                ble_central_get_max_request_payload_size(),
                ble_central_get_max_response_payload_size());
    }

    /* Perform key exchange if peripheral supports encryption */
    uint16_t cap_flags = ble_central_get_capability_flags();
    if (cap_flags & CAPABILITY_FLAG_ENCRYPTION_SUPPORTED) {
        LOG_INF("Peripheral supports encryption, performing key exchange...");
        err = ble_central_perform_key_exchange();
        if (err) {
            LOG_WRN("Key exchange failed (err %d), continuing without encryption", err);
        } else {
            LOG_INF("Encryption active: %s", ble_central_is_encrypted() ? "yes" : "no");
        }
    }

    /* Allow subscription to settle */
    k_sleep(K_MSEC(200));

    /* Run tests */
    int failures = 0;

    if (test_echo() != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_flash_read(MAX_TEST_PAYLOAD) != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_throughput() != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_data_write(MAX_TEST_PAYLOAD) != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_write_throughput() != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_counter_stream() != 0) {
        failures++;
    }

    k_sleep(K_MSEC(100));

    if (test_counter_upload() != 0) {
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
