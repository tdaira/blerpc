#include "ble_service.h"
#include <blerpc_protocol/container.h>
#include <blerpc_protocol/command.h>
#include "handlers.h"

#include <zephyr/kernel.h>
#include <zephyr/bluetooth/bluetooth.h>
#include <zephyr/bluetooth/gap.h>
#include <zephyr/logging/log.h>
#include <pb_encode.h>

#ifdef CONFIG_BLERPC_ENCRYPTION
#include <blerpc_protocol/crypto.h>
#include <mbedtls/platform_util.h>
#include <psa/crypto.h>
#endif

LOG_MODULE_REGISTER(ble_service, LOG_LEVEL_INF);

/* type(1) + name_len(1) + name(max 16) + data_len(2) */
#define CMD_HEADER_MAX_SIZE 20

static struct bt_uuid_128 blerpc_svc_uuid = BT_UUID_INIT_128(BLERPC_SERVICE_UUID);
static struct bt_uuid_128 blerpc_char_uuid = BT_UUID_INIT_128(BLERPC_CHAR_UUID);

static struct bt_conn *current_conn;
static struct container_assembler assembler;
static ble_service_stream_end_cb_t stream_end_cb;
static uint8_t transaction_counter;

#ifdef CONFIG_BLERPC_ENCRYPTION
static struct blerpc_crypto_session crypto_session;
static struct blerpc_peripheral_key_exchange peripheral_kx;
static bool encryption_active;

static int hex_to_bytes(const char *hex, uint8_t *out, size_t out_len)
{
    for (size_t i = 0; i < out_len; i++) {
        unsigned int byte;
        char tmp[3] = {hex[i * 2], hex[i * 2 + 1], '\0'};
        if (sscanf(tmp, "%02x", &byte) != 1) {
            return -1;
        }
        out[i] = (uint8_t)byte;
    }
    return 0;
}

static int load_keys(void)
{
    /* PSA Crypto must be initialized before any PSA operations */
    psa_status_t psa_rc = psa_crypto_init();
    if (psa_rc != PSA_SUCCESS) {
        LOG_ERR("psa_crypto_init failed: %d", (int)psa_rc);
        return -1;
    }

    const char *ed25519_hex = CONFIG_BLERPC_ED25519_PRIVATE_KEY;

    if (strlen(ed25519_hex) != 64) {
        LOG_ERR("Ed25519 key not configured (must be 64 hex chars)");
        return -1;
    }

    uint8_t ed25519_privkey[32];

    if (hex_to_bytes(ed25519_hex, ed25519_privkey, 32) != 0) {
        LOG_ERR("Invalid hex in Ed25519 key");
        return -1;
    }

    /* Initialize peripheral key exchange (X25519 keypair generated per session) */
    if (blerpc_peripheral_kx_init(&peripheral_kx, ed25519_privkey) != 0) {
        LOG_ERR("Failed to initialize peripheral key exchange");
        return -1;
    }

    mbedtls_platform_zeroize(ed25519_privkey, sizeof(ed25519_privkey));
    LOG_INF("Encryption keys loaded");
    return 0;
}
#endif /* CONFIG_BLERPC_ENCRYPTION */

/* Work queue for async response processing */
static struct k_work_q blerpc_work_q;
static K_THREAD_STACK_DEFINE(blerpc_work_stack, CONFIG_BLERPC_WORK_STACK_SIZE);

struct request_work {
    struct k_work work;
    uint8_t transaction_id;
    size_t len;
    uint8_t data[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE];
};

static struct request_work req_work;

/* ── Streaming container sender ──────────────────────────────────────── */

struct streaming_ctx {
    uint8_t transaction_id;
    uint16_t mtu;
    uint16_t total_length; /* total payload for FIRST container header */
    uint8_t buf[252];      /* one container at a time (max effective MTU) */
    uint8_t seq;
    uint8_t payload_used; /* payload bytes buffered in current container */
    bool first_sent;
    int error;
};

static int send_with_retry(const uint8_t *data, size_t len)
{
    int rc;
    for (int retries = 0; retries < 10; retries++) {
        rc = ble_service_notify(data, len);
        if (rc != -ENOMEM) {
            return rc;
        }
        k_sleep(K_MSEC(5));
    }
    LOG_ERR("Notify failed after retries: %d", rc);
    return rc;
}

static uint8_t streaming_header_size(struct streaming_ctx *ctx)
{
    return ctx->first_sent ? CONTAINER_SUBSEQUENT_HEADER_SIZE : CONTAINER_FIRST_HEADER_SIZE;
}

static uint8_t streaming_max_payload(struct streaming_ctx *ctx)
{
    return (uint8_t)((ctx->mtu - CONTAINER_ATT_OVERHEAD) - streaming_header_size(ctx));
}

static int streaming_flush_container(struct streaming_ctx *ctx)
{
    if (ctx->payload_used == 0) {
        return 0;
    }

    uint8_t hdr_size = streaming_header_size(ctx);

    /* Build container header in-place at buf[0..hdr_size-1] */
    ctx->buf[0] = ctx->transaction_id;
    ctx->buf[1] = ctx->seq;

    if (!ctx->first_sent) {
        ctx->buf[2] = ((CONTAINER_TYPE_FIRST & 0x03) << 6);
        ctx->buf[3] = (uint8_t)(ctx->total_length & 0xFF);
        ctx->buf[4] = (uint8_t)(ctx->total_length >> 8);
        ctx->buf[5] = ctx->payload_used;
    } else {
        ctx->buf[2] = ((CONTAINER_TYPE_SUBSEQUENT & 0x03) << 6);
        ctx->buf[3] = ctx->payload_used;
    }

    int rc = send_with_retry(ctx->buf, hdr_size + ctx->payload_used);
    if (rc < 0) {
        ctx->error = rc;
        return rc;
    }

    ctx->seq++;
    ctx->first_sent = true;
    ctx->payload_used = 0;
    return 0;
}

static int streaming_write(struct streaming_ctx *ctx, const uint8_t *data, size_t len)
{
    if (ctx->error) {
        return ctx->error;
    }

    while (len > 0) {
        uint8_t hdr_size = streaming_header_size(ctx);
        uint8_t max_payload = streaming_max_payload(ctx);
        uint8_t space = max_payload - ctx->payload_used;
        uint8_t n = (len < space) ? (uint8_t)len : space;

        memcpy(ctx->buf + hdr_size + ctx->payload_used, data, n);
        ctx->payload_used += n;
        data += n;
        len -= n;

        if (ctx->payload_used >= max_payload) {
            int rc = streaming_flush_container(ctx);
            if (rc < 0) {
                return rc;
            }
        }
    }
    return 0;
}

static bool streaming_pb_callback(pb_ostream_t *stream, const uint8_t *buf, size_t count)
{
    struct streaming_ctx *ctx = (struct streaming_ctx *)stream->state;
    return streaming_write(ctx, buf, count) == 0;
}

/* Callback for container_split_and_send */
static int container_send_cb(const uint8_t *data, size_t len, void *ctx)
{
    (void)ctx;
    return send_with_retry(data, len);
}

/* ── Request processing ──────────────────────────────────────────────── */

static void process_request(const uint8_t *data, size_t len, uint8_t transaction_id)
{
    /* Parse command */
    struct command_packet cmd;
    if (command_parse(data, len, &cmd) != 0) {
        LOG_ERR("Command parse failed");
        return;
    }

    if (cmd.cmd_type != COMMAND_TYPE_REQUEST) {
        LOG_ERR("Expected request, got type %d", cmd.cmd_type);
        return;
    }

    /* Look up handler */
    command_handler_fn handler = handlers_lookup(cmd.cmd_name, cmd.cmd_name_len);
    if (!handler) {
        LOG_ERR("Unknown command: %.*s", cmd.cmd_name_len, cmd.cmd_name);
        return;
    }

    /* Pass 1: Calculate protobuf encoded size (sizing stream, no I/O) */
    pb_ostream_t sizing = PB_OSTREAM_SIZING;
    int handler_rc = handler(cmd.data, cmd.data_len, &sizing);
    if (handler_rc == -2) {
        /* Handler manages its own response (e.g. stream handlers) */
        return;
    }
    if (handler_rc != 0) {
        LOG_ERR("Handler sizing pass failed");
        return;
    }
    size_t pb_size = sizing.bytes_written;

    /* Calculate total command payload size */
    size_t cmd_hdr_size = 2 + cmd.cmd_name_len + 2;
    size_t total_length = cmd_hdr_size + pb_size;

    /* Check response size against max */
    if (CONFIG_BLERPC_MAX_RESPONSE_PAYLOAD_SIZE < 65535 &&
        total_length > CONFIG_BLERPC_MAX_RESPONSE_PAYLOAD_SIZE) {
        uint8_t ctrl_buf[8];
        struct container_header ctrl = {
            .transaction_id = transaction_id,
            .sequence_number = 0,
            .type = CONTAINER_TYPE_CONTROL,
            .control_cmd = CONTROL_CMD_ERROR,
            .payload_len = 1,
        };
        uint8_t err_payload[1] = {BLERPC_ERROR_RESPONSE_TOO_LARGE};
        ctrl.payload = err_payload;
        int n = container_serialize(&ctrl, ctrl_buf, sizeof(ctrl_buf));
        if (n > 0) {
            send_with_retry(ctrl_buf, (size_t)n);
        }
        LOG_WRN("Response too large: %zu > %u", total_length,
                CONFIG_BLERPC_MAX_RESPONSE_PAYLOAD_SIZE);
        return;
    }

    uint8_t cmd_hdr[CMD_HEADER_MAX_SIZE];
    if (cmd_hdr_size > sizeof(cmd_hdr)) {
        LOG_ERR("Command name too long for response header: %u", cmd.cmd_name_len);
        return;
    }
    cmd_hdr[0] = (COMMAND_TYPE_RESPONSE & 0x01) << 7;
    cmd_hdr[1] = cmd.cmd_name_len;
    memcpy(cmd_hdr + 2, cmd.cmd_name, cmd.cmd_name_len);
    size_t dl_offset = 2 + cmd.cmd_name_len;
    cmd_hdr[dl_offset] = (uint8_t)(pb_size & 0xFF);
    cmd_hdr[dl_offset + 1] = (uint8_t)((pb_size >> 8) & 0xFF);

    /* Set up streaming container sender */
    uint16_t mtu = ble_service_get_mtu();

#ifdef CONFIG_BLERPC_ENCRYPTION
    if (encryption_active) {
        /* When encryption is active, we need to serialize the full command
         * payload first, encrypt it, then send via container splitter */
        static uint8_t cmd_plain_buf[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE];
        memcpy(cmd_plain_buf, cmd_hdr, cmd_hdr_size);

        /* Encode protobuf into the buffer after the command header */
        pb_ostream_t ostream = pb_ostream_from_buffer(cmd_plain_buf + cmd_hdr_size,
                                                      sizeof(cmd_plain_buf) - cmd_hdr_size);
        if (handler(cmd.data, cmd.data_len, &ostream) != 0) {
            LOG_ERR("Handler encode pass failed");
            return;
        }

        /* Encrypt the full command payload */
        static uint8_t
            encrypted_buf[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE + BLERPC_ENCRYPTED_OVERHEAD];
        size_t encrypted_len;
        if (blerpc_crypto_session_encrypt(&crypto_session, encrypted_buf, sizeof(encrypted_buf),
                                          &encrypted_len, cmd_plain_buf, total_length) != 0) {
            LOG_ERR("Response encryption failed");
            return;
        }

        /* Send encrypted payload via container splitter */
        int rc = container_split_and_send(transaction_id, encrypted_buf, encrypted_len, mtu,
                                          container_send_cb, NULL);
        if (rc < 0) {
            LOG_ERR("Encrypted container send failed: %d", rc);
        }
        return;
    }
#endif

    struct streaming_ctx sctx = {
        .transaction_id = transaction_id,
        .mtu = mtu,
        .total_length = (uint16_t)total_length,
    };

    /* Write command header into container stream */
    streaming_write(&sctx, cmd_hdr, cmd_hdr_size);

    /* Pass 2: Encode protobuf directly into container stream */
    pb_ostream_t ostream = {
        .callback = streaming_pb_callback,
        .state = &sctx,
        .max_size = SIZE_MAX,
        .bytes_written = 0,
    };

    if (handler(cmd.data, cmd.data_len, &ostream) != 0) {
        LOG_ERR("Handler encode pass failed");
        return;
    }

    /* Flush last partial container */
    streaming_flush_container(&sctx);

    if (sctx.error) {
        LOG_ERR("Streaming send failed: %d", sctx.error);
    }
}

static void request_work_handler(struct k_work *work)
{
    struct request_work *rw = CONTAINER_OF(work, struct request_work, work);
    process_request(rw->data, rw->len, rw->transaction_id);
}

/* ── BLE service ─────────────────────────────────────────────────────── */

static ssize_t on_write(struct bt_conn *conn, const struct bt_gatt_attr *attr, const void *buf,
                        uint16_t len, uint16_t offset, uint8_t flags)
{
    (void)attr;
    (void)offset;
    (void)flags;

    LOG_DBG("Write: %u bytes", len);

    struct container_header hdr;
    if (container_parse_header(buf, len, &hdr) != 0) {
        LOG_ERR("Container parse failed");
        return len;
    }

    /* Handle control containers inline (small, fast) */
    if (hdr.type == CONTAINER_TYPE_CONTROL) {
        if (hdr.control_cmd == CONTROL_CMD_TIMEOUT) {
            uint8_t ctrl_buf[8];
            struct container_header ctrl = {
                .transaction_id = hdr.transaction_id,
                .sequence_number = 0,
                .type = CONTAINER_TYPE_CONTROL,
                .control_cmd = CONTROL_CMD_TIMEOUT,
                .payload_len = 2,
            };
            uint8_t timeout_payload[2] = {
                (uint8_t)(CONFIG_BLERPC_TIMEOUT_MS & 0xFF),
                (uint8_t)(CONFIG_BLERPC_TIMEOUT_MS >> 8),
            };
            ctrl.payload = timeout_payload;
            int n = container_serialize(&ctrl, ctrl_buf, sizeof(ctrl_buf));
            if (n > 0) {
                ble_service_notify(ctrl_buf, (size_t)n);
            }
        } else if (hdr.control_cmd == CONTROL_CMD_STREAM_END_C2P) {
            if (stream_end_cb) {
                stream_end_cb(hdr.transaction_id);
            }
        } else if (hdr.control_cmd == CONTROL_CMD_CAPABILITIES) {
            uint8_t ctrl_buf[12];
            struct container_header ctrl = {
                .transaction_id = hdr.transaction_id,
                .sequence_number = 0,
                .type = CONTAINER_TYPE_CONTROL,
                .control_cmd = CONTROL_CMD_CAPABILITIES,
                .payload_len = 6,
            };
            uint8_t caps_payload[6];
            uint16_t max_req = CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE;
            uint16_t max_resp = CONFIG_BLERPC_MAX_RESPONSE_PAYLOAD_SIZE;
            uint16_t flags = 0;
#ifdef CONFIG_BLERPC_ENCRYPTION
            flags |= CAPABILITY_FLAG_ENCRYPTION_SUPPORTED;
#endif
            caps_payload[0] = (uint8_t)(max_req & 0xFF);
            caps_payload[1] = (uint8_t)(max_req >> 8);
            caps_payload[2] = (uint8_t)(max_resp & 0xFF);
            caps_payload[3] = (uint8_t)(max_resp >> 8);
            caps_payload[4] = (uint8_t)(flags & 0xFF);
            caps_payload[5] = (uint8_t)(flags >> 8);
            ctrl.payload = caps_payload;
            int n = container_serialize(&ctrl, ctrl_buf, sizeof(ctrl_buf));
            if (n > 0) {
                ble_service_notify(ctrl_buf, (size_t)n);
            }
#ifdef CONFIG_BLERPC_ENCRYPTION
        } else if (hdr.control_cmd == CONTROL_CMD_KEY_EXCHANGE) {
            /* Block KX re-initiation when encryption is already active */
            if (encryption_active) {
                LOG_WRN("Key exchange rejected: encryption already active");
                return len;
            }

            uint8_t kx_out[BLERPC_STEP2_SIZE]; /* large enough for step 2 or 4 */
            size_t kx_out_len;
            bool session_established;

            if (blerpc_peripheral_kx_handle_step(&peripheral_kx, hdr.payload, hdr.payload_len,
                                                 kx_out, sizeof(kx_out), &kx_out_len,
                                                 &crypto_session, &session_established) != 0) {
                LOG_ERR("Key exchange step processing failed");
                return len;
            }

            uint8_t resp_buf[BLERPC_STEP2_SIZE + CONTAINER_CONTROL_HEADER_SIZE];
            struct container_header kx_ctrl = {
                .transaction_id = hdr.transaction_id,
                .sequence_number = 0,
                .type = CONTAINER_TYPE_CONTROL,
                .control_cmd = CONTROL_CMD_KEY_EXCHANGE,
                .payload_len = kx_out_len,
                .payload = kx_out,
            };
            int n = container_serialize(&kx_ctrl, resp_buf, sizeof(resp_buf));
            if (n > 0) {
                send_with_retry(resp_buf, (size_t)n);
            }

            if (session_established) {
                encryption_active = true;
                LOG_INF("E2E encryption established");
            }
#endif /* CONFIG_BLERPC_ENCRYPTION */
        }
        return len;
    }

    /* Feed into assembler */
    int rc = container_assembler_feed(&assembler, &hdr);
    if (rc == 1) {
        /* Assembly complete — process via work queue to free BT RX thread */
        req_work.transaction_id = hdr.transaction_id;
#ifdef CONFIG_BLERPC_ENCRYPTION
        if (encryption_active) {
            /* Decrypt assembled payload (static to avoid stack overflow on BT RX thread) */
            static uint8_t decrypted[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE];
            size_t decrypted_len;
            if (blerpc_crypto_session_decrypt(&crypto_session, decrypted, sizeof(decrypted),
                                              &decrypted_len, assembler.buf,
                                              assembler.total_length) != 0) {
                LOG_ERR("Decryption failed");
                container_assembler_init(&assembler);
                return len;
            }
            req_work.len = decrypted_len;
            memcpy(req_work.data, decrypted, decrypted_len);
        } else {
            /* Reject unencrypted data when encryption is compiled in */
            LOG_WRN("Rejecting unencrypted payload (encryption enabled but not active)");
            container_assembler_init(&assembler);
            return len;
        }
#else
            req_work.len = assembler.total_length;
            memcpy(req_work.data, assembler.buf, assembler.total_length);
#endif
        container_assembler_init(&assembler);
        k_work_submit_to_queue(&blerpc_work_q, &req_work.work);
    } else if (rc < 0) {
        container_assembler_init(&assembler);
    }

    return len;
}

/* GATT service definition */
BT_GATT_SERVICE_DEFINE(blerpc_svc, BT_GATT_PRIMARY_SERVICE(&blerpc_svc_uuid),
                       BT_GATT_CHARACTERISTIC(&blerpc_char_uuid.uuid,
                                              BT_GATT_CHRC_WRITE_WITHOUT_RESP | BT_GATT_CHRC_NOTIFY,
                                              BT_GATT_PERM_WRITE, NULL, on_write, NULL),
                       BT_GATT_CCC(NULL, BT_GATT_PERM_READ | BT_GATT_PERM_WRITE), );

uint16_t ble_service_get_mtu(void)
{
    if (current_conn) {
        return bt_gatt_get_mtu(current_conn);
    }
    return 23; /* Default minimum */
}

int ble_service_notify(const uint8_t *data, size_t len)
{
    if (!current_conn) {
        return -ENOTCONN;
    }

    struct bt_gatt_notify_params params = {
        .attr = &blerpc_svc.attrs[2],
        .data = data,
        .len = len,
    };

    return bt_gatt_notify_cb(current_conn, &params);
}

static void connected(struct bt_conn *conn, uint8_t err)
{
    if (err) {
        LOG_ERR("Connection failed (err %u)", err);
        return;
    }
    LOG_INF("Connected");
    current_conn = bt_conn_ref(conn);
    container_assembler_init(&assembler);
    transaction_counter = 0;
#ifdef CONFIG_BLERPC_ENCRYPTION
    encryption_active = false;
    memset(&crypto_session, 0, sizeof(crypto_session));
    blerpc_peripheral_kx_reset(&peripheral_kx);
#endif
}

static const struct bt_data ad[] = {
    BT_DATA_BYTES(BT_DATA_FLAGS, (BT_LE_AD_GENERAL | BT_LE_AD_NO_BREDR)),
    BT_DATA_BYTES(BT_DATA_UUID128_ALL, BLERPC_SERVICE_UUID),
};

static const struct bt_data sd[] = {
    BT_DATA(BT_DATA_NAME_COMPLETE, CONFIG_BLERPC_DEVICE_NAME,
            sizeof(CONFIG_BLERPC_DEVICE_NAME) - 1),
};

static void disconnected(struct bt_conn *conn, uint8_t reason)
{
    LOG_INF("Disconnected (reason %u)", reason);
    if (current_conn) {
        bt_conn_unref(current_conn);
        current_conn = NULL;
    }
    container_assembler_init(&assembler);
#ifdef CONFIG_BLERPC_ENCRYPTION
    encryption_active = false;
    memset(&crypto_session, 0, sizeof(crypto_session));
    blerpc_peripheral_kx_reset(&peripheral_kx);
#endif

    int err = ble_service_start_advertising();
    if (err) {
        LOG_ERR("Failed to restart advertising (err %d)", err);
    }
}

int ble_service_start_advertising(void)
{
    return bt_le_adv_start(BT_LE_ADV_CONN, ad, ARRAY_SIZE(ad), sd, ARRAY_SIZE(sd));
}

BT_CONN_CB_DEFINE(conn_callbacks) = {
    .connected = connected,
    .disconnected = disconnected,
};

void ble_service_init(void)
{
    k_work_queue_init(&blerpc_work_q);
    k_work_queue_start(&blerpc_work_q, blerpc_work_stack, K_THREAD_STACK_SIZEOF(blerpc_work_stack),
                       K_PRIO_COOP(7), NULL);
    k_work_init(&req_work.work, request_work_handler);
    container_assembler_init(&assembler);

#ifdef CONFIG_BLERPC_ENCRYPTION
    if (load_keys() != 0) {
        LOG_WRN("Encryption keys not loaded — running without encryption");
    }
#endif
}

int ble_service_send_stream_end_p2c(uint8_t transaction_id)
{
    uint8_t ctrl_buf[8];
    struct container_header ctrl = {
        .transaction_id = transaction_id,
        .sequence_number = 0,
        .type = CONTAINER_TYPE_CONTROL,
        .control_cmd = CONTROL_CMD_STREAM_END_P2C,
        .payload_len = 0,
    };
    ctrl.payload = NULL;
    int n = container_serialize(&ctrl, ctrl_buf, sizeof(ctrl_buf));
    if (n < 0) {
        return -1;
    }
    return send_with_retry(ctrl_buf, (size_t)n);
}

void ble_service_set_stream_end_cb(ble_service_stream_end_cb_t cb)
{
    stream_end_cb = cb;
}

uint8_t ble_service_next_transaction_id(void)
{
    return transaction_counter++;
}

void ble_service_submit_work(struct k_work *work)
{
    k_work_submit_to_queue(&blerpc_work_q, work);
}

int ble_service_send_command_response(uint8_t transaction_id, const uint8_t *cmd_data,
                                      size_t cmd_len)
{
    uint16_t mtu = ble_service_get_mtu();

#ifdef CONFIG_BLERPC_ENCRYPTION
    if (encryption_active) {
        static uint8_t
            enc_buf[CONFIG_BLERPC_PROTOCOL_ASSEMBLER_BUF_SIZE + BLERPC_ENCRYPTED_OVERHEAD];
        size_t enc_len;
        if (blerpc_crypto_session_encrypt(&crypto_session, enc_buf, sizeof(enc_buf), &enc_len,
                                          cmd_data, cmd_len) != 0) {
            LOG_ERR("Stream response encryption failed");
            return -1;
        }
        return container_split_and_send(transaction_id, enc_buf, enc_len, mtu, container_send_cb,
                                        NULL);
    }
#endif

    return container_split_and_send(transaction_id, cmd_data, cmd_len, mtu, container_send_cb,
                                    NULL);
}
