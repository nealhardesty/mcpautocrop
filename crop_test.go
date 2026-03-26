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

	msg, err := AutoCrop(input, output, 0, 0)
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

	msg, err := AutoCrop(input, output, 0, 0)
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

	msg, err := AutoCrop(input, output, 0, 0)
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

	_, err := AutoCrop("/nonexistent/path/image.png", output, 0, 0)
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

	_, err := AutoCrop(input, output, 0, 0)
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

	msg, err := AutoCrop(input, output, 5, 0)
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

	msg, err := AutoCrop(input, output, 10, 0)
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

	box := findBoundingBox(img, img.Bounds(), white, 0)
	expected := image.Rect(5, 7, 6, 8)
	if box != expected {
		t.Errorf("expected %v, got %v", expected, box)
	}
}

func TestAutoCrop_Tolerance(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "test.png")
	output := filepath.Join(dir, "out.png")

	// Build a 50x50 image that mimics a real screenshot border: the top-left pixel
	// is pure white (255,255,255) but other border pixels vary slightly — exactly
	// the pattern seen in anti-aliased PNGs.
	//
	// With tolerance=0 (exact match against 255,255,255), the off-white border pixels
	// don't match and the bounding box spans the full image.
	// With tolerance=10, all near-white pixels are treated as background and only the
	// red subject remains.
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	pureWhite := color.RGBA{255, 255, 255, 255}
	offWhite := color.RGBA{252, 253, 252, 255} // distance ~3.5 from pure white
	red := color.RGBA{200, 50, 50, 255}

	// Fill with pure white first (this is what (0,0) will be).
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, pureWhite)
		}
	}
	// Scatter off-white pixels throughout the border area to simulate anti-aliasing.
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if (x+y)%3 == 1 { // every third pixel is slightly off
				img.Set(x, y, offWhite)
			}
		}
	}
	// Red subject at (10,10) 20x20 (overwrites any off-white there).
	for y := 10; y < 30; y++ {
		for x := 10; x < 30; x++ {
			img.Set(x, y, red)
		}
	}

	f, err := os.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	// tolerance=0: exact match against (253,253,253); off-white border pixels look like
	// subject → bounding box == full image → "already optimally cropped".
	msg, err := AutoCrop(input, output, 0, 0)
	if err != nil {
		t.Fatalf("AutoCrop returned error: %v", err)
	}
	if !strings.Contains(msg, "already optimally cropped") {
		t.Errorf("expected no-op with tolerance=0, got: %q", msg)
	}

	// tolerance=10: off-white pixels are within range of the background → subject isolated.
	msg, err = AutoCrop(input, output, 0, 10)
	if err != nil {
		t.Fatalf("AutoCrop returned error: %v", err)
	}
	if !strings.HasPrefix(msg, "Successfully cropped") {
		t.Fatalf("expected crop with tolerance=10, got: %q", msg)
	}
	cf, err := os.Open(output)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()
	cropped, err := png.Decode(cf)
	if err != nil {
		t.Fatal(err)
	}
	b := cropped.Bounds()
	if b.Dx() != 20 || b.Dy() != 20 {
		t.Errorf("expected 20x20 crop with tolerance=10, got %dx%d", b.Dx(), b.Dy())
	}
}
