#!/usr/bin/env python3
"""Read SEGGER RTT output from a target board via pylink.

Usage:
    python3 tools/rtt_reader.py [--reset] [--rtt-address 0x20000090]

Connects to the first available J-Link, reads RTT channel 0, and prints
output. Press Ctrl+C to stop.

Tries the RTTERMINAL API first (fast, works with updated J-Link FW).
Falls back to direct memory reads if RTTERMINAL fails (works with old FW).

The RTT address can be found with:
    arm-zephyr-eabi-nm zephyr.elf | grep _SEGGER_RTT
"""

import argparse
import struct
import sys
import time

import pylink


def read_u32(jlink, addr):
    data = jlink.memory_read(addr, 4)
    return struct.unpack("<I", bytes(data))[0]


def run_rtterminal(jlink, rtt_addr):
    """Use RTTERMINAL API (requires J-Link FW support)."""
    jlink.rtt_start(block_address=rtt_addr)

    for _ in range(50):
        try:
            num_up = jlink.rtt_get_num_up_buffers()
            if num_up > 0:
                print(f"RTT started via RTTERMINAL API ({num_up} up buffers)", file=sys.stderr)
                break
        except pylink.errors.JLinkRTTException:
            pass
        time.sleep(0.1)
    else:
        jlink.rtt_stop()
        return False

    try:
        while True:
            data = jlink.rtt_read(0, 1024)
            if data:
                sys.stdout.write(bytes(data).decode("utf-8", errors="replace"))
                sys.stdout.flush()
            else:
                time.sleep(0.05)
    except KeyboardInterrupt:
        print("\nStopped.", file=sys.stderr)
    finally:
        jlink.rtt_stop()
    return True


def run_memory_read(jlink, rtt_addr):
    """Read RTT ring buffer directly from target memory."""
    magic = bytes(jlink.memory_read(rtt_addr, 10))
    if magic != b"SEGGER RTT":
        print(f"ERROR: No RTT magic at 0x{rtt_addr:08x} (got {magic!r})", file=sys.stderr)
        return False

    # Parse aUp[0] ring buffer descriptor (starts at rtt_addr + 24)
    # struct { const char *sName; char *pBuffer; uint SizeOfBuffer; uint WrOff; uint RdOff; uint Flags; }
    UP0 = rtt_addr + 24
    buf_ptr = read_u32(jlink, UP0 + 4)
    buf_size = read_u32(jlink, UP0 + 8)
    print(f"RTT started via memory read (buf=0x{buf_ptr:08x} size={buf_size})", file=sys.stderr)

    # Set RdOff = WrOff to skip old data
    wr_off = read_u32(jlink, UP0 + 12)
    jlink.memory_write(UP0 + 16, list(struct.pack("<I", wr_off)))

    try:
        while True:
            wr_off = read_u32(jlink, UP0 + 12)
            rd_off = read_u32(jlink, UP0 + 16)

            if wr_off == rd_off:
                time.sleep(0.05)
                continue

            if wr_off > rd_off:
                data = bytes(jlink.memory_read(buf_ptr + rd_off, wr_off - rd_off))
            else:
                data = bytes(jlink.memory_read(buf_ptr + rd_off, buf_size - rd_off))
                if wr_off > 0:
                    data += bytes(jlink.memory_read(buf_ptr, wr_off))

            jlink.memory_write(UP0 + 16, list(struct.pack("<I", wr_off)))
            sys.stdout.write(data.decode("utf-8", errors="replace"))
            sys.stdout.flush()

    except KeyboardInterrupt:
        print("\nStopped.", file=sys.stderr)
    return True


def main():
    parser = argparse.ArgumentParser(description="Read RTT from target board")
    parser.add_argument("--reset", action="store_true", help="Reset target before reading")
    parser.add_argument("--rtt-address", type=lambda x: int(x, 0), default=0x20000090,
                        help="RTT control block address (from nm zephyr.elf | grep _SEGGER_RTT)")
    parser.add_argument("--device", default="EFR32BG22C224F512IM40",
                        help="J-Link device name (default: EFR32BG22C224F512IM40)")
    parser.add_argument("--memory-only", action="store_true",
                        help="Skip RTTERMINAL API, use memory reads only")
    args = parser.parse_args()

    jlink = pylink.JLink()
    jlink.open()
    jlink.set_tif(pylink.enums.JLinkInterfaces.SWD)
    jlink.connect(args.device, speed=4000)

    if args.reset:
        jlink.reset(halt=False)
        time.sleep(0.5)

    try:
        if not args.memory_only:
            if run_rtterminal(jlink, args.rtt_address):
                return
            print("RTTERMINAL API failed, falling back to memory read", file=sys.stderr)

        if not run_memory_read(jlink, args.rtt_address):
            sys.exit(1)
    finally:
        jlink.close()


if __name__ == "__main__":
    main()
