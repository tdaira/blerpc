#include "ble_central.h"
#include <blerpc_protocol/container.h>

#include <zephyr/kernel.h>
#include <zephyr/bluetooth/bluetooth.h>
#include <zephyr/bluetooth/conn.h>
#include <zephyr/bluetooth/gatt.h>
#include <zephyr/bluetooth/uuid.h>
#include <zephyr/logging/log.h>

LOG_MODULE_REGISTER(ble_central, LOG_LEVEL_INF);

/* UUIDs */
static struct bt_uuid_128 blerpc_svc_uuid = BT_UUID_INIT_128(BLERPC_SERVICE_UUID);
static struct bt_uuid_128 blerpc_char_uuid = BT_UUID_INIT_128(BLERPC_CHAR_UUID);

/* Connection state */
static struct bt_conn *current_conn;
static uint16_t char_value_handle;
static struct bt_gatt_subscribe_params subscribe_params;

/* Container assembler for incoming notifications */
static struct container_assembler assembler;

/* Response callback */
static ble_central_response_cb_t response_cb;

/* Synchronization semaphores */
static K_SEM_DEFINE(connect_sem, 0, 1);
static K_SEM_DEFINE(discover_sem, 0, 1);
static K_SEM_DEFINE(mtu_sem, 0, 1);

/* ── Scan callbacks ──────────────────────────────────────────────────── */

static void device_found(const bt_addr_le_t *addr, int8_t rssi, uint8_t type,
                         struct net_buf_simple *ad)
{
    /* Only look at connectable advertisements */
    if (type != BT_GAP_ADV_TYPE_ADV_IND && type != BT_GAP_ADV_TYPE_ADV_DIRECT_IND) {
        return;
    }

    /* Check advertisement data for device name "blerpc" */
    struct net_buf_simple_state state;
    net_buf_simple_save(ad, &state);

    while (ad->len > 1) {
        uint8_t field_len = net_buf_simple_pull_u8(ad);
        if (field_len == 0 || field_len > ad->len) {
            break;
        }
        uint8_t field_type = net_buf_simple_pull_u8(ad);
        field_len--; /* Exclude type byte */

        if ((field_type == BT_DATA_NAME_COMPLETE || field_type == BT_DATA_NAME_SHORTENED) &&
            field_len == 6 && memcmp(ad->data, "blerpc", 6) == 0) {
            /* Found it */
            char addr_str[BT_ADDR_LE_STR_LEN];
            bt_addr_le_to_str(addr, addr_str, sizeof(addr_str));
            LOG_INF("Found 'blerpc' device: %s (RSSI %d)", addr_str, rssi);

            int err = bt_le_scan_stop();
            if (err) {
                LOG_ERR("Scan stop failed (err %d)", err);
                net_buf_simple_restore(ad, &state);
                return;
            }

            struct bt_conn_le_create_param create_param = BT_CONN_LE_CREATE_PARAM_INIT(
                BT_CONN_LE_OPT_NONE, BT_GAP_SCAN_FAST_INTERVAL, BT_GAP_SCAN_FAST_WINDOW);

            struct bt_le_conn_param conn_param = BT_LE_CONN_PARAM_INIT(12, 24, 0, 100);

            err = bt_conn_le_create(addr, &create_param, &conn_param, &current_conn);
            if (err) {
                LOG_ERR("Create connection failed (err %d)", err);
            }

            net_buf_simple_restore(ad, &state);
            return;
        }

        net_buf_simple_pull(ad, field_len);
    }

    net_buf_simple_restore(ad, &state);
}

/* ── GATT discovery ──────────────────────────────────────────────────── */

static struct bt_gatt_discover_params discover_params;
static uint16_t svc_start_handle;
static uint16_t svc_end_handle;

static uint8_t discover_desc_cb(struct bt_conn *conn, const struct bt_gatt_attr *attr,
                                struct bt_gatt_discover_params *params)
{
    if (!attr) {
        LOG_INF("Descriptor discovery complete");
        k_sem_give(&discover_sem);
        return BT_GATT_ITER_STOP;
    }

    LOG_INF("Descriptor found: handle %u", attr->handle);

    /* Subscribe for notifications */
    subscribe_params.notify = NULL; /* Set later */
    subscribe_params.value_handle = char_value_handle;
    subscribe_params.ccc_handle = attr->handle;
    subscribe_params.value = BT_GATT_CCC_NOTIFY;

    k_sem_give(&discover_sem);
    return BT_GATT_ITER_STOP;
}

static uint8_t discover_char_cb(struct bt_conn *conn, const struct bt_gatt_attr *attr,
                                struct bt_gatt_discover_params *params)
{
    if (!attr) {
        LOG_ERR("Characteristic not found");
        k_sem_give(&discover_sem);
        return BT_GATT_ITER_STOP;
    }

    struct bt_gatt_chrc *chrc = (struct bt_gatt_chrc *)attr->user_data;
    char_value_handle = chrc->value_handle;
    LOG_INF("Characteristic found: value_handle %u", char_value_handle);

    k_sem_give(&discover_sem);
    return BT_GATT_ITER_STOP;
}

static uint8_t discover_svc_cb(struct bt_conn *conn, const struct bt_gatt_attr *attr,
                               struct bt_gatt_discover_params *params)
{
    if (!attr) {
        LOG_ERR("Service not found");
        k_sem_give(&discover_sem);
        return BT_GATT_ITER_STOP;
    }

    struct bt_gatt_service_val *svc = (struct bt_gatt_service_val *)attr->user_data;
    svc_start_handle = attr->handle;
    svc_end_handle = svc->end_handle;
    LOG_INF("Service found: handles %u-%u", svc_start_handle, svc_end_handle);

    k_sem_give(&discover_sem);
    return BT_GATT_ITER_STOP;
}

static int gatt_discover(void)
{
    int err;

    /* Phase 1: Discover primary service */
    memset(&discover_params, 0, sizeof(discover_params));
    discover_params.uuid = &blerpc_svc_uuid.uuid;
    discover_params.func = discover_svc_cb;
    discover_params.start_handle = BT_ATT_FIRST_ATTRIBUTE_HANDLE;
    discover_params.end_handle = BT_ATT_LAST_ATTRIBUTE_HANDLE;
    discover_params.type = BT_GATT_DISCOVER_PRIMARY;

    err = bt_gatt_discover(current_conn, &discover_params);
    if (err) {
        LOG_ERR("Service discover failed (err %d)", err);
        return err;
    }
    k_sem_take(&discover_sem, K_FOREVER);

    if (svc_start_handle == 0) {
        LOG_ERR("Service not found");
        return -ENOENT;
    }

    /* Phase 2: Discover characteristic */
    memset(&discover_params, 0, sizeof(discover_params));
    discover_params.uuid = &blerpc_char_uuid.uuid;
    discover_params.func = discover_char_cb;
    discover_params.start_handle = svc_start_handle;
    discover_params.end_handle = svc_end_handle;
    discover_params.type = BT_GATT_DISCOVER_CHARACTERISTIC;

    err = bt_gatt_discover(current_conn, &discover_params);
    if (err) {
        LOG_ERR("Char discover failed (err %d)", err);
        return err;
    }
    k_sem_take(&discover_sem, K_FOREVER);

    if (char_value_handle == 0) {
        LOG_ERR("Characteristic not found");
        return -ENOENT;
    }

    /* Phase 3: Discover CCC descriptor */
    memset(&discover_params, 0, sizeof(discover_params));
    discover_params.uuid = BT_UUID_GATT_CCC;
    discover_params.func = discover_desc_cb;
    discover_params.start_handle = char_value_handle + 1;
    discover_params.end_handle = svc_end_handle;
    discover_params.type = BT_GATT_DISCOVER_DESCRIPTOR;

    err = bt_gatt_discover(current_conn, &discover_params);
    if (err) {
        LOG_ERR("Descriptor discover failed (err %d)", err);
        return err;
    }
    k_sem_take(&discover_sem, K_FOREVER);

    return 0;
}

/* ── Notification handler ────────────────────────────────────────────── */

static uint8_t notify_handler(struct bt_conn *conn, struct bt_gatt_subscribe_params *params,
                              const void *data, uint16_t length)
{
    if (!data) {
        LOG_INF("Notifications disabled");
        params->value_handle = 0;
        return BT_GATT_ITER_STOP;
    }

    LOG_DBG("Notification: %u bytes", length);

    struct container_header hdr;
    if (container_parse_header(data, length, &hdr) != 0) {
        LOG_ERR("Container parse failed");
        return BT_GATT_ITER_CONTINUE;
    }

    int rc = container_assembler_feed(&assembler, &hdr);
    if (rc == 1) {
        /* Assembly complete */
        if (response_cb) {
            response_cb(assembler.buf, assembler.total_length);
        }
        container_assembler_init(&assembler);
    } else if (rc < 0) {
        LOG_ERR("Assembler error");
        container_assembler_init(&assembler);
    }

    return BT_GATT_ITER_CONTINUE;
}

/* ── Connection callbacks ────────────────────────────────────────────── */

static void connected_cb(struct bt_conn *conn, uint8_t err)
{
    if (err) {
        LOG_ERR("Connection failed (err %u)", err);
        if (current_conn) {
            bt_conn_unref(current_conn);
            current_conn = NULL;
        }
        k_sem_give(&connect_sem);
        return;
    }

    LOG_INF("Connected");
    k_sem_give(&connect_sem);
}

static void disconnected_cb(struct bt_conn *conn, uint8_t reason)
{
    LOG_INF("Disconnected (reason %u)", reason);
    if (current_conn) {
        bt_conn_unref(current_conn);
        current_conn = NULL;
    }
}

BT_CONN_CB_DEFINE(conn_callbacks) = {
    .connected = connected_cb,
    .disconnected = disconnected_cb,
};

/* ── MTU exchange ────────────────────────────────────────────────────── */

static struct bt_gatt_exchange_params mtu_exchange_params;

static void mtu_exchange_cb(struct bt_conn *conn, uint8_t err,
                            struct bt_gatt_exchange_params *params)
{
    if (err) {
        LOG_ERR("MTU exchange failed (err %u)", err);
    } else {
        LOG_INF("MTU exchanged: %u", bt_gatt_get_mtu(conn));
    }
    k_sem_give(&mtu_sem);
}

/* ── Public API ──────────────────────────────────────────────────────── */

void ble_central_init(ble_central_response_cb_t cb)
{
    response_cb = cb;
    container_assembler_init(&assembler);
}

int ble_central_connect(void)
{
    int err;

    LOG_INF("Scanning for 'blerpc' peripheral...");

    err = bt_le_scan_start(BT_LE_SCAN_ACTIVE, device_found);
    if (err) {
        LOG_ERR("Scan start failed (err %d)", err);
        return err;
    }

    /* Wait for connection */
    k_sem_take(&connect_sem, K_FOREVER);

    if (!current_conn) {
        LOG_ERR("Connection failed");
        return -ENOTCONN;
    }

    /* Request data length update */
    struct bt_conn_le_data_len_param dl_param = {
        .tx_max_len = 251,
        .tx_max_time = 2120,
    };
    err = bt_conn_le_data_len_update(current_conn, &dl_param);
    if (err) {
        LOG_WRN("Data length update failed (err %d), continuing", err);
    }

    /* Exchange MTU */
    mtu_exchange_params.func = mtu_exchange_cb;
    err = bt_gatt_exchange_mtu(current_conn, &mtu_exchange_params);
    if (err) {
        LOG_ERR("MTU exchange request failed (err %d)", err);
    } else {
        k_sem_take(&mtu_sem, K_SECONDS(5));
    }

    /* GATT discovery */
    err = gatt_discover();
    if (err) {
        return err;
    }

    /* Subscribe for notifications */
    subscribe_params.notify = notify_handler;
    err = bt_gatt_subscribe(current_conn, &subscribe_params);
    if (err) {
        LOG_ERR("Subscribe failed (err %d)", err);
        return err;
    }

    LOG_INF("Subscribed to notifications");
    return 0;
}

int ble_central_write(const uint8_t *data, size_t len)
{
    if (!current_conn) {
        return -ENOTCONN;
    }

    return bt_gatt_write_without_response(current_conn, char_value_handle, data, len, false);
}

uint16_t ble_central_get_mtu(void)
{
    if (current_conn) {
        return bt_gatt_get_mtu(current_conn);
    }
    return 23;
}
