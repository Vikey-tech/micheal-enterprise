"""Validated, bounded image conversion called synchronously by the Go upload handler."""
import os
import sys
import warnings
from pathlib import Path
from PIL import Image, ImageOps

Image.MAX_IMAGE_PIXELS = 24_000_000

def optimize(source, target):
    source, target = Path(source), Path(target)
    if source.stat().st_size > 10 * 1024 * 1024:
        raise ValueError("Image exceeds 10 MB")
    with warnings.catch_warnings():
        warnings.simplefilter("error", Image.DecompressionBombWarning)
        with Image.open(source) as image:
            if image.format not in {"JPEG", "PNG", "WEBP"}:
                raise ValueError("Unsupported image type")
            if image.width * image.height > Image.MAX_IMAGE_PIXELS:
                raise ValueError("Image exceeds 24 megapixels")
            image.verify()
        with Image.open(source) as image:
            image = ImageOps.exif_transpose(image).convert("RGB")
            image.thumbnail((1920, 1440))
            # Re-encode pixels to strip EXIF, embedded content and original metadata.
            clean = Image.new("RGB", image.size)
            clean.paste(image)
            temporary = target.with_suffix(".tmp")
            try:
                clean.save(temporary, format="WEBP", quality=85, method=4)
                os.replace(temporary, target)
            finally:
                temporary.unlink(missing_ok=True)
    return str(target)

if __name__ == "__main__":
    optimize(sys.argv[1], sys.argv[2])
