# Firmware Build, Flash and Debug Guide

This guide covers building, flashing, and viewing logs for the C firmware
targets (peripheral and central\_fw) on supported boards.

## Prerequisites

- [nRF Connect SDK](https://developer.nordicsemi.com/nRF_Connect_SDK/) v2.9.0 installed at `~/ncs`
- [Zephyr SDK](https://github.com/zephyrproject-rtos/sdk-ng) v0.17.0 installed at `~/zephyr-sdk-0.17.0`
- `west` build tool (installed via pip)
- SEGGER J-Link software (for RTT and flashing)
- A supported board connected via USB

## Supported Boards

| Board | Role | RAM | Notes |
|-------|------|-----|-------|
| nRF54L15 DK | Peripheral, Central | 188 KB | Full feature support |
| EFR32xG22E EK2710A | Peripheral only | 32 KB | Flash disabled to fit RTT logging |

## Environment Setup

Set these environment variables before running build commands:

```bash
export ZEPHYR_BASE=~/ncs/zephyr
export ZEPHYR_SDK_INSTALL_DIR=~/zephyr-sdk-0.17.0
export ZEPHYR_TOOLCHAIN_VARIANT=zephyr
WEST=/Library/Frameworks/Python.framework/Versions/3.13/bin/west
```

## Building

All build commands should be run from the project root (`blerpc/`).

### Peripheral Firmware

**nRF54L15 DK:**

```bash
$WEST -z ~/ncs/zephyr build -d peripheral_fw/build \
  -b nrf54l15dk/nrf54l15/cpuapp peripheral_fw \
  -- -DNCS_TOOLCHAIN_VERSION=NONE
```

**EFR32xG22E EK:**

```bash
$WEST -z ~/ncs/zephyr build -d peripheral_fw/build_xg22 \
  -b xg22_ek2710a peripheral_fw \
  -- -DNCS_TOOLCHAIN_VERSION=NONE -DBOARD_ROOT=$(pwd)
```

> The `-DBOARD_ROOT` flag is required because the xg22\_ek2710a board
> definition is in the project tree (`boards/silabs/xg22_ek2710a/`),
> not in the nRF Connect SDK.

### Central Firmware

**nRF54L15 DK:**

```bash
$WEST -z ~/ncs/zephyr build -d central_fw/build \
  -b nrf54l15dk/nrf54l15/cpuapp central_fw \
  -- -DNCS_TOOLCHAIN_VERSION=NONE
```

> Central firmware is only supported on nRF54L15 (EFR32xG22E does not
> have enough RAM).

## Flashing

### nRF54L15 DK

Uses `nrfutil` runner (auto-detected by west):

```bash
$WEST -z ~/ncs/zephyr flash -d peripheral_fw/build
```

or for central:

```bash
$WEST -z ~/ncs/zephyr flash -d central_fw/build
```

If the board is not detected, check that:
- The USB cable is connected to the J-Link port (not the DK VCOM port)
- SWD is enabled via nRF Connect for Desktop &rarr; Board Configurator

### EFR32xG22E EK

Uses `jlink` runner (auto-detected by west):

```bash
$WEST -z ~/ncs/zephyr flash -d peripheral_fw/build_xg22
```

### Resetting Without Reflashing

**nRF54L15:**

```bash
nrfjprog --reset
```

**EFR32xG22E (via JLink Commander):**

```bash
JLinkExe -Device EFR32BG22C224F512IM40 -If SWD -Speed 4000 -AutoConnect 1
# In JLink prompt:
r
g
q
```

## Viewing Logs

### nRF54L15 DK (RTT)

The nRF54L15 firmware has logging enabled via SEGGER RTT. Use
`JLinkRTTLogger` or `JLinkRTTClient` to view live output.

**JLinkRTTLogger** (captures to file):

```bash
JLinkRTTLogger -Device NRF54L15_M33 -If SWD -Speed 4000 -RTTChannel 0 /tmp/rtt.log
```

> Use `NRF54L15_M33` as the device name, not `NRF54L15_XXAA`.

**JLinkRTTClient** (interactive terminal):

```bash
JLinkRTTClient -Device NRF54L15_M33 -If SWD -Speed 4000
```

**Note:** The RTT buffer is 1024 bytes by default. If the firmware
produces a lot of output (e.g., throughput tests), earlier messages may
be overwritten before you can read them. To capture from boot, start
the RTT logger before resetting the board.

### EFR32xG22E EK (RTT via memory read)

The EFR32xG22E board has only 32 KB RAM. To fit RTT logging, flash
support is disabled and buffer sizes are reduced. The board config
(`peripheral_fw/boards/xg22_ek2710a.conf`) enables `CONFIG_LOG_MODE_MINIMAL`
which routes `LOG_INF` etc. through `printk` to an RTT console with a
128-byte buffer.

> **Note:** `JLinkRTTLogger` does not work with this board because
> the RTT Control Block is not 1024-byte aligned. Use the
> `tools/rtt_reader.py` script instead, which connects via pylink
> (RTTERMINAL API with memory-read fallback).
>
> **Prerequisite:** Update the on-board J-Link adapter firmware via
> Simplicity Studio 5 (Launcher &rarr; Adapter FW &rarr; Update).
> The RTTERMINAL API requires updated firmware; older firmware falls
> back to slower direct memory reads.

**Reading RTT logs:**

```bash
# Terminal 1: Start RTT reader (requires pylink: pip install pylink-square)
python3 tools/rtt_reader.py

# Or with board reset to capture boot messages:
python3 tools/rtt_reader.py --reset
```

The default RTT address (`0x20000090`) matches the current build.
If the address changes after code modifications, find the new address:

```bash
~/zephyr-sdk-0.17.0/arm-zephyr-eabi/bin/arm-zephyr-eabi-nm \
  peripheral_fw/build_xg22/peripheral_fw/zephyr/zephyr.elf | grep _SEGGER_RTT
```

Then pass it: `python3 tools/rtt_reader.py --rtt-address 0xNEWADDR`

**Running tests with RTT:**

```bash
# Terminal 1: Start RTT reader
python3 tools/rtt_reader.py

# Terminal 2: Run integration tests
cd central_py && python3 -m pytest tests/test_integration.py -v -s
```

> **Important:** Flash reads are disabled in the RTT-enabled config
> (`CONFIG_FLASH=n`) to free RAM. Flash-related tests will be skipped.
> To restore flash support, disable RTT in the board config.

### Central Firmware Logs (nRF54L15)

The central firmware uses both RTT and UART backends. View RTT logs the
same way as the peripheral:

```bash
JLinkRTTLogger -Device NRF54L15_M33 -If SWD -Speed 4000 -RTTChannel 0 /tmp/rtt_central.log
```

UART output is available on the DK's virtual COM port (typically
`/dev/tty.usbmodem*` on macOS):

```bash
screen /dev/tty.usbmodem14101 115200
```

## Running Integration Tests

Integration tests require a peripheral board running the firmware and a
Mac with BLE to act as the Python Central.

```bash
# Flash the peripheral first, then:
cd central_py
python3 -m pytest tests/test_integration.py -v -s
```

For role-reversal tests (C Central + Python Peripheral):

```bash
# Terminal 1: Start Python peripheral
python3 peripheral_py/server.py

# Terminal 2: Flash and run central firmware
$WEST -z ~/ncs/zephyr flash -d central_fw/build
# View results via RTT
JLinkRTTLogger -Device NRF54L15_M33 -If SWD -Speed 4000 -RTTChannel 0 /tmp/rtt.log
```

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `west flash` hangs or fails | SWD disabled on nRF54L15 | Enable via Board Configurator |
| RTT shows no output | Board already finished tests | Reset board after starting RTT logger |
| EFR32 build: RAM overflow | Feature too large for 32 KB | Reduce buffer sizes or disable logging |
| `ENOMEM` in RTT logs | BLE TX buffers full | Retry logic is built in; increase `BT_L2CAP_TX_BUF_COUNT` if persistent |
| Flash fails on EFR32 | Wrong J-Link target | Ensure J-Link is connected to EFR32 board, not nRF |
