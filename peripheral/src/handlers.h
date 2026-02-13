#ifndef BLERPC_HANDLERS_H
#define BLERPC_HANDLERS_H

#include "generated_handlers.h"

/**
 * Initialize stream handlers (register STREAM_END_C2P callback).
 * Call after ble_service_init().
 */
void handlers_stream_init(void);

#endif /* BLERPC_HANDLERS_H */
