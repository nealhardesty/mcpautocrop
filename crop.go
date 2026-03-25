package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// subImager is satisfied by images that support cropping via SubImage.
type subImager interface {
	SubImage(r image.Rectangle) image.Image
}

// AutoCrop reads the image at inputPath, crops away the background color (sampled
// from the top-left pixel), and writes the result to outputPath.
// border adds extra padding (in pixels) around the detected subject, clamped to
// the image edges. Pass 0 for a tight crop with no padding.
// It returns a human-readable status message on success or an error.
func AutoCrop(inputPath, outputPath string, border int) (string, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return "", fmt.Errorf("Error: Input file '%s' could not be opened.", inputPath)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("Error: image decoding failed. Ensure file is a valid PNG or JPEG.")
	}

	bounds := img.Bounds()
	bgColor := img.At(bounds.Min.X, bounds.Min.Y)

	box := findBoundingBox(img, bounds, bgColor)
	if box.Empty() {
		return "Image consists entirely of background color; cannot crop.", nil
	}

	// Expand by requested border, clamped to the original image bounds.
	if border > 0 {
		box = image.Rect(
			max(bounds.Min.X, box.Min.X-border),
			max(bounds.Min.Y, box.Min.Y-border),
			min(bounds.Max.X, box.Max.X+border),
			min(bounds.Max.Y, box.Max.Y+border),
		)
	}

	// Check whether the (possibly padded) box would change anything.
	if box == bounds {
		return "Image is already optimally cropped. No changes made.", nil
	}

	si, ok := img.(subImager)
	if !ok {
		return "", fmt.Errorf("Error: image type does not support cropping.")
	}
	cropped := si.SubImage(box)

	if err := saveImage(cropped, outputPath); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully cropped and saved to %s", outputPath), nil
}

// findBoundingBox returns the smallest rectangle containing all pixels that
// differ from bgColor.
func findBoundingBox(img image.Image, bounds image.Rectangle, bgColor color.Color) image.Rectangle {
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	bgR, bgG, bgB, bgA := bgColor.RGBA()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if r != bgR || g != bgG || b != bgB || a != bgA {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if minX > maxX || minY > maxY {
		return image.Rectangle{}
	}

	// maxX/maxY are inclusive pixel coordinates; image.Rectangle is exclusive on Max.
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

// saveImage writes img to path, choosing the encoder based on the file extension.
func saveImage(img image.Image, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error: Cannot write to '%s'. Permission denied.", path)
	}
	defer out.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(out, img, nil); err != nil {
			return fmt.Errorf("Error: failed to encode JPEG: %w", err)
		}
	default:
		// Default to PNG for .png and any unknown extension.
		if err := png.Encode(out, img); err != nil {
			return fmt.Errorf("Error: failed to encode PNG: %w", err)
		}
	}
	return nil
}
