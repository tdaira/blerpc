#include <zephyr/kernel.h>
#include <zephyr/bluetooth/bluetooth.h>
#include <zephyr/logging/log.h>
#ifdef CONFIG_BLERPC_ENCRYPTION
#include <mbedtls/platform.h>
#endif

#include "ble_service.h"
#include "handlers.h"

LOG_MODULE_REGISTER(main, LOG_LEVEL_INF);

int main(void)
{
    int err;

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

    ble_service_init();
    handlers_stream_init();

    err = ble_service_start_advertising();
    if (err) {
        LOG_ERR("Advertising failed to start (err %d)", err);
        return err;
    }
    LOG_INF("Advertising started");

    return 0;
}
