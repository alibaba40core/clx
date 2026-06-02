#!/usr/bin/env python3
"""Render terminal capture text files as PNG screenshots."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


def load_font(size: int) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    candidates = [
        Path(r"C:\Windows\Fonts\Consola.ttf"),
        Path(r"C:\Windows\Fonts\cour.ttf"),
        Path("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"),
    ]
    for path in candidates:
        if path.is_file():
            return ImageFont.truetype(str(path), size=size)
    return ImageFont.load_default()


def render_terminal(text: str, title: str, out_path: Path, font_size: int = 16) -> None:
    font = load_font(font_size)
    lines = text.replace("\r\n", "\n").replace("\r", "\n").split("\n")
    pad_x, pad_y = 24, 20
    line_h = font_size + 6
    char_w = font_size * 0.6
    max_cols = max((len(line) for line in lines), default=40)
    width = int(pad_x * 2 + max_cols * char_w)
    height = int(pad_y * 2 + len(lines) * line_h + 36)

    img = Image.new("RGB", (max(width, 640), max(height, 120)), "#1e1e1e")
    draw = ImageDraw.Draw(img)

    # Title bar
    draw.rectangle([0, 0, img.width, 32], fill="#323232")
    draw.ellipse([12, 10, 22, 20], fill="#ff5f57")
    draw.ellipse([28, 10, 38, 20], fill="#febc2e")
    draw.ellipse([44, 10, 54, 20], fill="#28c840")
    draw.text((64, 8), title, fill="#cccccc", font=load_font(14))

    y = pad_y + 24
    for line in lines:
        if line.startswith("PS ") or line.startswith("C:\\") or line.startswith("/"):
            color = "#4ec9b0"
        elif line.startswith("Intent:") or line.startswith("Source:") or line.startswith("Command:"):
            color = "#9cdcfe"
        elif line.startswith("Risk:"):
            color = "#dcdcaa"
        elif "error" in line.lower() or "fail" in line.lower():
            color = "#f48771"
        else:
            color = "#d4d4d4"
        draw.text((pad_x, y), line, fill=color, font=font)
        y += line_h

    out_path.parent.mkdir(parents=True, exist_ok=True)
    img.save(out_path, "PNG")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("input_txt", type=Path)
    parser.add_argument("output_png", type=Path)
    parser.add_argument("--title", default="Terminal")
    args = parser.parse_args()
    text = args.input_txt.read_text(encoding="utf-8", errors="replace")
    render_terminal(text, args.title, args.output_png)
    print(f"wrote {args.output_png}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
