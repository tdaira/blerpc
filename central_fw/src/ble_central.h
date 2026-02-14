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
 * Callback for received error control containers.
 */
typedef void (*ble_central_error_cb_t)(uint8_t error_code);

/**
 * Callback for STREAM_END_P2C control container.
 */
typedef void (*ble_central_stream_end_cb_t)(void);

/**
 * Initialize the BLE central module.
 * @param resp_cb  Callback invoked when a complete response is assembled
 * @param err_cb   Callback invoked when an ERROR control container is received
 */
void ble_central_init(ble_central_response_cb_t resp_cb, ble_central_error_cb_t err_cb);

/**
 * Set callback for STREAM_END_P2C reception.
 */
void ble_central_set_stream_end_cb(ble_central_stream_end_cb_t cb);

/**
 * Send a STREAM_END_C2P control container to peripheral.
 * @return 0 on success, negative on error
 */
int ble_central_send_stream_end_c2p(void);

/**
 * Scan for and connect to a device advertising the given name.
 * Blocks until connected and GATT discovery + subscription complete.
 * @param device_name  Name to match in advertisement data
 * @return 0 on success, negative on error
 */
int ble_central_connect(const char *device_name);

/**
 * Send data to the peripheral (write without response).
 * @return 0 on success, negative on error
 */
int ble_central_write(const uint8_t *data, size_t len);

/**
 * Get the current connection MTU.
 */
uint16_t ble_central_get_mtu(void);

/**
 * Request capabilities from the peripheral.
 * Blocks until response received or timeout.
 * @return 0 on success, negative on error/timeout
 */
int ble_central_request_capabilities(void);

/**
 * Get peripheral's max request payload size (0 if unknown).
 */
uint16_t ble_central_get_max_request_payload_size(void);

/**
 * Get peripheral's max response payload size (0 if unknown).
 */
uint16_t ble_central_get_max_response_payload_size(void);

#ifdef __cplusplus
}
#endif

#endif /* BLERPC_BLE_CENTRAL_H */
