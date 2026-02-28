#!/usr/bin/env python3
"""Generate app icons for all 4 bleRPC Central mobile clients.

Requires: pip install Pillow
"""

import json
import os
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

# Root of the monorepo
ROOT = Path(__file__).resolve().parent.parent

# Colors (blerpc.net dark theme)
BG_COLOR = (0x1A, 0x1B, 0x26)
ACCENT_BLUE = (0x00, 0x82, 0xFC)
TEXT_LIGHT = (0xC0, 0xCA, 0xF5)
TEXT_SECONDARY = (0xA9, 0xB1, 0xD6)

MASTER_SIZE = 1024

# Android mipmap sizes
ANDROID_SIZES = {
    "mipmap-mdpi": 48,
    "mipmap-hdpi": 72,
    "mipmap-xhdpi": 96,
    "mipmap-xxhdpi": 144,
    "mipmap-xxxhdpi": 192,
}

# iOS icon sizes (pixel sizes)
IOS_SIZES = [40, 58, 60, 80, 87, 120, 152, 167, 180, 1024]

# Platform configs: (subtitle, android_res_dir or None, ios_appiconset_dir or None)
PLATFORMS = [
    (
        "Kotlin",
        ROOT / "central_android/app/src/main/res",
        None,
    ),
    (
        "Swift",
        None,
        ROOT / "central_ios/BlerpcCentral/Assets.xcassets/AppIcon.appiconset",
    ),
    (
        "Flutter",
        ROOT / "central_flutter/android/app/src/main/res",
        ROOT / "central_flutter/ios/Runner/Assets.xcassets/AppIcon.appiconset",
    ),
    (
        "React Native",
        ROOT / "central_rn/android/app/src/main/res",
        ROOT / "central_rn/ios/BlerpcCentral/Images.xcassets/AppIcon.appiconset",
    ),
]


def try_load_font(size: int) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    """Try to load a bold font, falling back gracefully."""
    font_paths = [
        "/System/Library/Fonts/Helvetica.ttc",
        "/System/Library/Fonts/SFNSMono.ttf",
        "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
        "/usr/share/fonts/TTF/DejaVuSans-Bold.ttf",
    ]
    for path in font_paths:
        if os.path.exists(path):
            try:
                return ImageFont.truetype(path, size)
            except Exception:
                continue
    return ImageFont.load_default()


def render_icon(subtitle: str) -> Image.Image:
    """Render a 1024x1024 icon with bleRPC branding and platform subtitle."""
    img = Image.new("RGB", (MASTER_SIZE, MASTER_SIZE), BG_COLOR)
    draw = ImageDraw.Draw(img)

    # Font sizes relative to 1024
    main_font_size = int(MASTER_SIZE * 0.22)
    sub_font_size = int(MASTER_SIZE * 0.10)

    main_font = try_load_font(main_font_size)
    sub_font = try_load_font(sub_font_size)

    # Measure "ble" and "RPC" separately
    ble_text = "ble"
    rpc_text = "RPC"

    ble_bbox = draw.textbbox((0, 0), ble_text, font=main_font)
    rpc_bbox = draw.textbbox((0, 0), rpc_text, font=main_font)
    ble_w = ble_bbox[2] - ble_bbox[0]
    rpc_w = rpc_bbox[2] - rpc_bbox[0]
    main_h = max(ble_bbox[3] - ble_bbox[1], rpc_bbox[3] - rpc_bbox[1])

    total_w = ble_w + rpc_w

    # Subtitle
    sub_bbox = draw.textbbox((0, 0), subtitle, font=sub_font)
    sub_w = sub_bbox[2] - sub_bbox[0]
    sub_h = sub_bbox[3] - sub_bbox[1]

    # Vertical layout: center both lines as a group
    spacing = int(MASTER_SIZE * 0.04)
    group_h = main_h + spacing + sub_h
    y_start = (MASTER_SIZE - group_h) // 2

    # Draw "ble" in accent blue
    x_start = (MASTER_SIZE - total_w) // 2
    draw.text((x_start, y_start), ble_text, fill=ACCENT_BLUE, font=main_font)
    # Draw "RPC" in light
    draw.text((x_start + ble_w, y_start), rpc_text, fill=TEXT_LIGHT, font=main_font)

    # Draw subtitle
    sub_x = (MASTER_SIZE - sub_w) // 2
    sub_y = y_start + main_h + spacing
    draw.text((sub_x, sub_y), subtitle, fill=TEXT_SECONDARY, font=sub_font)

    return img


def save_android_icons(master: Image.Image, res_dir: Path) -> None:
    """Save Android mipmap icons."""
    for folder, size in ANDROID_SIZES.items():
        out_dir = res_dir / folder
        out_dir.mkdir(parents=True, exist_ok=True)
        resized = master.resize((size, size), Image.LANCZOS)
        resized.save(out_dir / "ic_launcher.png")
        print(f"  Android: {out_dir / 'ic_launcher.png'} ({size}x{size})")


def save_ios_icons(master: Image.Image, appiconset_dir: Path) -> None:
    """Save iOS icons and Contents.json."""
    appiconset_dir.mkdir(parents=True, exist_ok=True)

    # Define the icon set entries
    images = []
    filenames_written = set()

    # iPhone icons
    iphone_entries = [
        ("20x20", "2x", 40),
        ("20x20", "3x", 60),
        ("29x29", "2x", 58),
        ("29x29", "3x", 87),
        ("40x40", "2x", 80),
        ("40x40", "3x", 120),
        ("60x60", "2x", 120),
        ("60x60", "3x", 180),
    ]

    for size_str, scale, px in iphone_entries:
        filename = f"icon-{px}.png"
        images.append({
            "size": size_str,
            "idiom": "iphone",
            "filename": filename,
            "scale": scale,
        })
        if filename not in filenames_written:
            resized = master.resize((px, px), Image.LANCZOS)
            resized.save(appiconset_dir / filename)
            filenames_written.add(filename)

    # iPad icons
    ipad_entries = [
        ("20x20", "1x", 20),
        ("20x20", "2x", 40),
        ("29x29", "1x", 29),
        ("29x29", "2x", 58),
        ("40x40", "1x", 40),
        ("40x40", "2x", 80),
        ("76x76", "1x", 76),
        ("76x76", "2x", 152),
        ("83.5x83.5", "2x", 167),
    ]

    for size_str, scale, px in ipad_entries:
        filename = f"icon-{px}.png"
        images.append({
            "size": size_str,
            "idiom": "ipad",
            "filename": filename,
            "scale": scale,
        })
        if filename not in filenames_written:
            resized = master.resize((px, px), Image.LANCZOS)
            resized.save(appiconset_dir / filename)
            filenames_written.add(filename)

    # App Store icon
    filename = "icon-1024.png"
    images.append({
        "size": "1024x1024",
        "idiom": "ios-marketing",
        "filename": filename,
        "scale": "1x",
    })
    master.save(appiconset_dir / filename)
    filenames_written.add(filename)

    contents = {
        "images": images,
        "info": {"version": 1, "author": "xcode"},
    }
    with open(appiconset_dir / "Contents.json", "w") as f:
        json.dump(contents, f, indent=2)
        f.write("\n")

    print(f"  iOS: {appiconset_dir} ({len(filenames_written)} PNGs + Contents.json)")


def main() -> None:
    for subtitle, android_dir, ios_dir in PLATFORMS:
        print(f"\n[{subtitle}]")
        master = render_icon(subtitle)

        if android_dir is not None:
            save_android_icons(master, android_dir)
        if ios_dir is not None:
            save_ios_icons(master, ios_dir)

    print("\nDone!")


if __name__ == "__main__":
    main()
