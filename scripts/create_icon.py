"""Create a multi-resolution Windows icon from the official dwrean.net favicon.
Fallback: the dwrean.net logo from the publisher's dwrean-radio repository.
Pillow is only required during the build, not to run the application.
"""
from io import BytesIO
from pathlib import Path
from urllib.request import Request, urlopen
from PIL import Image, ImageOps

OUT = Path("assets/dwrean-cleaner.ico")
OUT.parent.mkdir(parents=True, exist_ok=True)
SOURCES = [
    ("dwrean.net favicon", "https://www.dwrean.net/favicon.ico"),
    ("dwrean.net publisher logo", "https://raw.githubusercontent.com/dwrean-net/dwrean-radio/0a5ea9efab3e791b5196a9acb8d5f3da5e194f14/assets/dwrean-logo.png"),
]
last_error = None
for name, url in SOURCES:
    try:
        req = Request(url, headers={"User-Agent": "dwrean-cleaner-build/1.0"})
        with urlopen(req, timeout=20) as response:
            blob = response.read(600_000)
        with Image.open(BytesIO(blob)) as im:
            im.seek(0)
            src = im.convert("RGBA")
        if min(src.size) < 16:
            raise ValueError("Brand image is too small to be used as an icon")
        canvas = Image.new("RGBA", (256, 256), (255, 255, 255, 0))
        thumb = ImageOps.contain(src, (238, 238), Image.Resampling.LANCZOS)
        canvas.alpha_composite(thumb, ((256-thumb.width)//2, (256-thumb.height)//2))
        canvas.save(OUT, format="ICO", sizes=[(16,16), (24,24), (32,32), (48,48), (64,64), (128,128), (256,256)])
        print(f"Icon generated from {name} ({src.width}x{src.height}).")
        break
    except Exception as exc:
        last_error = exc
        print(f"Source unavailable: {name}: {exc}")
else:
    raise SystemExit(f"Cannot create the branding icon: {last_error}")
