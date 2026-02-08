#ifndef BLERPC_BLE_CENTRAL_H
#define BLERPC_BLE_CENTRAL_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* blerpc Service UUID: 12340001-0000-1000-8000-00805f9b34fb */
#define BLERPC_SERVICE_UUID BT_UUID_128_ENCODE(0x12340001, 0x0000, 0x1000, 0x8000, 0x00805f9b34fb)

/* blerpc Characteristic UUID: 12340002-0000-1000-8000-00805f9b34fb */
#define BLERPC_CHAR_UUID BT_UUID_128_ENCODE(0x12340002, 0x0000, 0x1000, 0x8000, 0x00805f9b34fb)

/**
 * Callback for received RPC response data (assembled payload).
 */
typedef void (*ble_central_response_cb_t)(const uint8_t *data, size_t len);

/**
 * Initialize the BLE central module.
 * @param cb  Callback invoked when a complete response is assembled
 */
void ble_central_init(ble_central_response_cb_t cb);

/**
 * Scan for and connect to a device advertising name "blerpc".
 * Blocks until connected and GATT discovery + subscription complete.
 * @return 0 on success, negative on error
 */
int ble_central_connect(void);

/**
 * Send data to the peripheral (write without response).
 * @return 0 on success, negative on error
 */
int ble_central_write(const uint8_t *data, size_t len);

/**
 * Get the current connection MTU.
 */
uint16_t ble_central_get_mtu(void);

#ifdef __cplusplus
}
#endif

#endif /* BLERPC_BLE_CENTRAL_H */
