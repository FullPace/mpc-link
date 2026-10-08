# Renders the logo SVGs to PNGs (transparent) with headless Chromium: python3 render.py
import pathlib
from playwright.sync_api import sync_playwright
here = pathlib.Path(__file__).parent
HEAD = "44 0 144 113"   # the head with its headphone band, without the "d" and "b": legible at tray sizes
jobs = [  # (svg, out, width, height, pad: fraction of the box around the logo, viewBox or None for the whole logo)
    ("black.svg", "tray-mac.png", 72, 44, 0.04, None),         # menu bar, 2x of 36x22 pt: the whole logo
    ("red.svg", "tray-win.png", 64, 64, 0.02, HEAD),           # Windows tray, square
    ("red.svg", "appicon-1024.png", 1024, 1024, 0.10, HEAD),   # app icon source
]
with sync_playwright() as p:
    b = p.chromium.launch()
    for svg, out, w, h, pad, vb in jobs:
        page = b.new_page(viewport={"width": w, "height": h})
        src = (here / svg).read_text()
        if vb:
            src = src.replace('viewBox="0 0 232.3 113.18"', 'viewBox="%s"' % vb, 1)
        page.set_content(f'<html><body style="margin:0;background:transparent;display:flex;align-items:center;justify-content:center;width:{w}px;height:{h}px">'
                         f'<div style="width:{w*(1-2*pad)}px;height:{h*(1-2*pad)}px;display:flex;align-items:center;justify-content:center">'
                         + src.replace("<svg ", '<svg style="max-width:100%;max-height:100%;width:100%;height:100%" preserveAspectRatio="xMidYMid meet" ', 1)
                         + "</div></body></html>")
        page.screenshot(path=str(here / out), omit_background=True)
        print(out)
    b.close()
