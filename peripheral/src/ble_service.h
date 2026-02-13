#ifndef BLERPC_BLE_SERVICE_H
#define BLERPC_BLE_SERVICE_H

#include <zephyr/bluetooth/bluetooth.h>
#include <zephyr/bluetooth/gatt.h>

#ifdef __cplusplus
extern "C" {
#endif

/* blerpc Service UUID: 12340001-0000-1000-8000-00805f9b34fb */
#define BLERPC_SERVICE_UUID BT_UUID_128_ENCODE(0x12340001, 0x0000, 0x1000, 0x8000, 0x00805f9b34fb)

/* blerpc Characteristic UUID: 12340002-0000-1000-8000-00805f9b34fb */
#define BLERPC_CHAR_UUID BT_UUID_128_ENCODE(0x12340002, 0x0000, 0x1000, 0x8000, 0x00805f9b34fb)

/**
 * Initialize the BLE service (work queue, assembler).
 * Call after bt_enable() but before starting advertising.
 */
void ble_service_init(void);

/**
 * Start BLE advertising.
 * @return 0 on success, negative on error
 */
int ble_service_start_advertising(void);

/**
 * Get the current connection's MTU.
 */
uint16_t ble_service_get_mtu(void);

/**
 * Send a notification to the connected Central.
 * @param data  Data to send
 * @param len   Length of data
 * @return 0 on success, negative on error
 */
int ble_service_notify(const uint8_t *data, size_t len);

/**
 * Send a STREAM_END_P2C control container.
 * @param transaction_id Transaction ID
 * @return 0 on success, negative on error
 */
int ble_service_send_stream_end_p2c(uint8_t transaction_id);

/**
 * Callback type for stream end notification (C->P).
 * Called when STREAM_END_C2P is received from Central.
 */
typedef void (*ble_service_stream_end_cb_t)(uint8_t transaction_id);

/**
 * Register a callback for STREAM_END_C2P reception.
 */
void ble_service_set_stream_end_cb(ble_service_stream_end_cb_t cb);

/**
 * Get the next transaction ID (incrementing counter).
 */
uint8_t ble_service_next_transaction_id(void);

#ifdef __cplusplus
}
#endif

#endif /* BLERPC_BLE_SERVICE_H */
