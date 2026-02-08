cmake_minimum_required(VERSION 3.20.0)

# blerpc-protocol library (container + command layers)
set(BLERPC_PROTOCOL_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../../blerpc-protocol/c
    CACHE PATH "Path to blerpc-protocol C library")
list(APPEND ZEPHYR_EXTRA_MODULES ${BLERPC_PROTOCOL_DIR})

find_package(Zephyr REQUIRED HINTS $ENV{ZEPHYR_BASE})
project(blerpc_peripheral)

# nanopb
list(APPEND CMAKE_MODULE_PATH ${ZEPHYR_BASE}/../modules/lib/nanopb/extra)
include(FindNanopb)

target_sources(app PRIVATE
    src/main.c
    src/ble_service.c
    src/handlers.c
    src/generated_handlers.c
    src/blerpc.pb.c
)

target_include_directories(app PRIVATE
    src
)
