import tempfile
import unittest
from pathlib import Path
from PIL import Image
from optimize_images import optimize

class OptimizeTests(unittest.TestCase):
    def test_resize_and_remove_metadata(self):
        with tempfile.TemporaryDirectory() as directory:
            source, target = Path(directory)/"photo.jpg", Path(directory)/"result.webp"
            exif = Image.Exif(); exif[270] = "private metadata"
            Image.new("RGB", (3000, 2000), "brown").save(source, exif=exif)
            optimize(source, target)
            with Image.open(target) as image:
                self.assertEqual(image.format, "WEBP")
                self.assertEqual(image.size, (1920, 1280))
                self.assertFalse(image.getexif())
    def test_reject_text_and_gif(self):
        with tempfile.TemporaryDirectory() as directory:
            source, target = Path(directory)/"bad", Path(directory)/"out.webp"
            source.write_text("<script>alert(1)</script>")
            with self.assertRaises(Exception): optimize(source, target)
            Image.new("RGB", (10,10)).save(source, format="GIF")
            with self.assertRaises(ValueError): optimize(source, target)
            self.assertFalse(target.exists())
    def test_reject_oversized_file(self):
        with tempfile.TemporaryDirectory() as directory:
            source = Path(directory)/"large"
            with source.open("wb") as file: file.truncate(10 * 1024 * 1024 + 1)
            with self.assertRaises(ValueError): optimize(source, Path(directory)/"out.webp")

if __name__ == "__main__": unittest.main()
