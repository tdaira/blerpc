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
)

# Generate the nanopb message code (blerpc.pb.c/.h) from the proto at build
# time so it always matches proto/blerpc.proto and proto/blerpc.options.
# Resolve to an absolute path (no "..") so protoc does not report the input as
# shadowed in --proto_path.
get_filename_component(BLERPC_PROTO_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../proto ABSOLUTE)
# The nanopb runtime is already provided by CONFIG_NANOPB; keep standalone mode
# off so nanopb_generate_cpp does not add a duplicate runtime (multiple-definition).
set(NANOPB_GENERATE_CPP_STANDALONE OFF)
nanopb_generate_cpp(proto_srcs proto_hdrs
    RELPATH ${BLERPC_PROTO_DIR}
    ${BLERPC_PROTO_DIR}/blerpc.proto)
target_sources(app PRIVATE ${proto_srcs} ${proto_hdrs})

target_include_directories(app PRIVATE
    src
    ${CMAKE_CURRENT_BINARY_DIR}
)
