# Renders the square MPC Link logo (SVG/black.svg; white.svg is its inverse) to the app's PNGs with headless Chromium, keeping it square.
import pathlib
from playwright.sync_api import sync_playwright
from PIL import Image
here = pathlib.Path(__file__).parent
def shot(page, svg, size, out):
    page.set_viewport_size({"width": size, "height": size})
    src = (here / "SVG" / svg).read_text().replace("<svg ", f'<svg width="{size}" height="{size}" ', 1)
    page.set_content(f'<html><body style="margin:0;background:transparent">{src}</body></html>')
    page.screenshot(path=str(here / out), omit_background=True, clip={"x": 0, "y": 0, "width": size, "height": size})
with sync_playwright() as p:
    b = p.chromium.launch(); page = b.new_page()
    shot(page, "black.svg", 44, "tray-mac-src.png")       # menu bar: 22x22 pt at 2x
    shot(page, "black.svg", 64, "tray-win.png")           # Windows tray
    shot(page, "black.svg", 1024, "appicon-1024.png")     # app icon source
    b.close()
# template image: the black body stays, the white parts (screen, pads) become see-through
im = Image.open(here / "tray-mac-src.png").convert("RGBA")
px = im.load()
for y in range(im.size[1]):
    for x in range(im.size[0]):
        r, g, b_, a = px[x, y]
        lum = (r * 30 + g * 59 + b_ * 11) // 100
        px[x, y] = (0, 0, 0, a * (255 - lum) // 255)
im.save(here / "tray-mac.png")
c = Image.new("RGBA", (420, 300), (236, 236, 236, 255))
for i, s in enumerate((22, 44)):
    c.alpha_composite(Image.open(here / "tray-mac.png").resize((s, s), Image.LANCZOS), (10 + i * 60, 10))
d = Image.new("RGBA", (130, 64), (40, 40, 42, 255)); c.alpha_composite(d, (130, 0))
w = Image.open(here / "tray-mac.png"); wd = Image.new("RGBA", w.size, (255, 255, 255, 0)); wd.putalpha(w.split()[3])
c.alpha_composite(wd.resize((22, 22), Image.LANCZOS), (150, 20)); c.alpha_composite(wd, (190, 10))
for i, s in enumerate((16, 32, 64)):
    c.alpha_composite(Image.open(here / "tray-win.png").resize((s, s), Image.LANCZOS), (10 + i * 80, 90))
c.alpha_composite(Image.open(here / "appicon-1024.png").resize((128, 128), Image.LANCZOS), (280, 90))
c.save(here / "preview.png")
print("ok")
