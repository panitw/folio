# Usage: uv run --with zxing-cpp --with pillow python3 tools/codescan/scan.py <pdf>
#
# Independent decode check for folio8's code elements: rasterises every filled
# rectangle ("re f") in a PDF's content streams and prints each barcode or QR
# code zxing-cpp finds, with its format, error-correction level and text.
import re
import sys
import zlib

import zxingcpp
from PIL import Image, ImageDraw


def rects_of(pdf):
    rects = []
    for m in re.finditer(rb'stream\r?\n(.*?)\r?\nendstream', pdf, re.S):
        body = m.group(1)
        try:
            body = zlib.decompress(body)
        except Exception:
            pass
        for x, y, w, h in re.findall(rb'([-\d.]+) ([-\d.]+) ([-\d.]+) ([-\d.]+) re f', body):
            rects.append(tuple(float(v) for v in (x, y, w, h)))
    return rects


def main():
    if len(sys.argv) != 2:
        sys.exit('usage: scan.py <pdf>')
    rects = rects_of(open(sys.argv[1], 'rb').read())
    print('rects', len(rects))
    if not rects:
        sys.exit(1)
    scale = 8
    minx = min(r[0] for r in rects) - 20
    maxx = max(r[0] + r[2] for r in rects) + 20
    miny = min(r[1] for r in rects) - 20
    maxy = max(r[1] + r[3] for r in rects) + 20
    img = Image.new('L', (int((maxx - minx) * scale), int((maxy - miny) * scale)), 255)
    draw = ImageDraw.Draw(img)
    for x, y, w, h in rects:
        draw.rectangle([(x - minx) * scale, (maxy - (y + h)) * scale, (x + w - minx) * scale - 1, (maxy - y) * scale - 1], fill=0)
    results = zxingcpp.read_barcodes(img)
    for r in results:
        print('format', r.format, 'ec', r.ec_level or '-', 'text', repr(r.bytes.decode('utf-8', 'replace')))
    print(len(results), 'decoded')


if __name__ == '__main__':
    main()
