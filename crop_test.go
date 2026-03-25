package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestPNG writes a PNG to path. The image is (width x height) pixels
// filled with bgColor, with a rectangular subject of subjectColor painted
// at (sx, sy)-(sx+sw, sy+sh).
func createTestPNG(t *testing.T, path string, width, height int, bgColor, subjectColor color.RGBA, sx, sy, sw, sh int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill background.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}
	// Paint subject.
	for y := sy; y < sy+sh; y++ {
		for x := sx; x < sx+sw; x++ {
			img.Set(x, y, subjectColor)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("createTestPNG: os.Create: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("createTestPNG: png.Encode: %v", err)
	}
}

func TestAutoCrop_CropsToSubject(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "test.png")
	output := filepath.Join(dir, "out.png")

	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{200, 50, 50, 255}

	// 100x100 image, white background, red 20x30 subject at (10, 15).
	createTestPNG(t, input, 100, 100, white, red, 10, 15, 20, 30)

	msg, err := AutoCrop(input, output, 0)
	if err != nil {
		t.Fatalf("AutoCrop returned error: %v", err)
	}
	if !strings.HasPrefix(msg, "Successfully cropped") {
		t.Fatalf("unexpected message: %q", msg)
	}

	// Verify the output image has the expected dimensions.
	f, err := os.Open(output)
	if err != nil {
		t.Fatalf("could not open output: %v", err)
	}
	defer f.Close()
	outImg, err := png.Decode(f)
	if err != nil {
		t.Fatalf("could not decode output: %v", err)
	}

	bounds := outImg.Bounds()
	if bounds.Dx() != 20 || bounds.Dy() != 30 {
		t.Errorf("expected 20x30 crop, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestAutoCrop_AlreadyOptimal(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "test.png")
	output := filepath.Join(dir, "out.png")

	// Build an image that is already optimally cropped:
	// - (0,0) is white (background)
	// - every other pixel is red (subject)
	// The bounding box spans the full image, so no cropping is needed.
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{200, 50, 50, 255}
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, red)
		}
	}
	img.Set(0, 0, white) // background anchor

	f, err := os.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	msg, err := AutoCrop(input, output, 0)
	if err != nil {
		t.Fatalf("AutoCrop returned error: %v", err)
	}
	if !strings.Contains(msg, "already optimally cropped") {
		t.Errorf("expected 'already optimally cropped' message, got: %q", msg)
	}
}

func TestAutoCrop_EntirelyBackground(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "test.png")
	output := filepath.Join(dir, "out.png")

	white := color.RGBA{255, 255, 255, 255}
	// Fill entire image with the same color as the background pixel.
	createTestPNG(t, input, 50, 50, white, white, 0, 0, 50, 50)

	msg, err := AutoCrop(input, output, 0)
	if err != nil {
		t.Fatalf("AutoCrop returned error: %v", err)
	}
	if !strings.Contains(msg, "entirely of background color") {
		t.Errorf("expected 'entirely of background color' message, got: %q", msg)
	}
}

func TestAutoCrop_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out.png")

	_, err := AutoCrop("/nonexistent/path/image.png", output, 0)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "could not be opened") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAutoCrop_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "corrupt.png")
	output := filepath.Join(dir, "out.png")

	// Write garbage bytes.
	if err := os.WriteFile(input, []byte("not an image"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := AutoCrop(input, output, 0)
	if err == nil {
		t.Fatal("expected error for corrupt file, got nil")
	}
	if !strings.Contains(err.Error(), "image decoding failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAutoCrop_Border(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "test.png")
	output := filepath.Join(dir, "out.png")

	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{200, 50, 50, 255}

	// 100x100 image, white background, red 20x30 subject at (10, 15).
	// Tight crop → 20x30. With border=5 → 30x40 (subject+5 on each side).
	createTestPNG(t, input, 100, 100, white, red, 10, 15, 20, 30)

	msg, err := AutoCrop(input, output, 5)
	if err != nil {
		t.Fatalf("AutoCrop returned error: %v", err)
	}
	if !strings.HasPrefix(msg, "Successfully cropped") {
		t.Fatalf("unexpected message: %q", msg)
	}

	f, err := os.Open(output)
	if err != nil {
		t.Fatalf("could not open output: %v", err)
	}
	defer f.Close()
	outImg, err := png.Decode(f)
	if err != nil {
		t.Fatalf("could not decode output: %v", err)
	}

	bounds := outImg.Bounds()
	if bounds.Dx() != 30 || bounds.Dy() != 40 {
		t.Errorf("expected 30x40 crop with border=5, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestAutoCrop_BorderClampedToImageEdge(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "test.png")
	output := filepath.Join(dir, "out.png")

	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{200, 50, 50, 255}

	// 100x100 image, subject at (2,2) 96x96.
	// Tight box: (2,2,98,98). With border=10 → expands to (0,0,100,100) = full bounds.
	// Clamping kicks in on all four sides, and box == original bounds → "already optimally cropped".
	createTestPNG(t, input, 100, 100, white, red, 2, 2, 96, 96)

	msg, err := AutoCrop(input, output, 10)
	if err != nil {
		t.Fatalf("AutoCrop returned error: %v", err)
	}
	if !strings.Contains(msg, "already optimally cropped") {
		t.Errorf("expected 'already optimally cropped' when border fills image, got: %q", msg)
	}
}

func TestFindBoundingBox_SinglePixelSubject(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{200, 50, 50, 255}

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, white)
		}
	}
	img.Set(5, 7, red)

	box := findBoundingBox(img, img.Bounds(), white)
	expected := image.Rect(5, 7, 6, 8)
	if box != expected {
		t.Errorf("expected %v, got %v", expected, box)
	}
}
