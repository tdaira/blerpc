cmake_minimum_required(VERSION 3.20.0)
find_package(Zephyr REQUIRED HINTS $ENV{ZEPHYR_BASE})
project(blerpc_central)

# nanopb
list(APPEND CMAKE_MODULE_PATH ${ZEPHYR_BASE}/../modules/lib/nanopb/extra)
include(FindNanopb)

target_sources(app PRIVATE
    src/main.c
    src/ble_central.c
    src/container.c
    src/command.c
    src/blerpc.pb.c
)

target_include_directories(app PRIVATE
    src
)

target_compile_definitions(app PRIVATE
    CONTAINER_ASSEMBLER_BUF_SIZE=12288
)
