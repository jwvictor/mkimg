package effects

import (
	"image"
	"image/color"
	"testing"
)

// createTestImage makes a solid-color image for testing.
func createTestImage(w, h int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

func TestApplyUnknownFilter(t *testing.T) {
	img := createTestImage(10, 10, color.NRGBA{128, 128, 128, 255})
	_, err := Apply(img, "faketown", nil)
	if err == nil {
		t.Error("expected error for unknown filter")
	}
}

func TestApplyAllFilters(t *testing.T) {
	// Verify every known filter can be applied without error
	img := createTestImage(20, 20, color.NRGBA{128, 64, 200, 255})

	filters := []struct {
		name   string
		params map[string]float64
	}{
		{"blur", map[string]float64{"radius": 2}},
		{"sharpen", map[string]float64{"radius": 1}},
		{"brightness", map[string]float64{"value": 20}},
		{"contrast", map[string]float64{"value": 10}},
		{"saturation", map[string]float64{"value": 30}},
		{"gamma", map[string]float64{"value": 1.5}},
		{"hue", map[string]float64{"value": 90}},
		{"grayscale", nil},
		{"sepia", nil},
		{"invert", nil},
		{"pixelate", map[string]float64{"size": 5}},
		{"vignette", map[string]float64{"strength": 0.5}},
		{"noise", map[string]float64{"amount": 15}},
		{"posterize", map[string]float64{"levels": 4}},
		{"emboss", nil},
		{"edge", nil},
		{"glow", map[string]float64{"radius": 3, "strength": 0.3}},
		{"duotone", map[string]float64{"r1": 0, "g1": 0, "b1": 50, "r2": 255, "g2": 200, "b2": 100}},
	}

	for _, f := range filters {
		result, err := Apply(img, f.name, f.params)
		if err != nil {
			t.Errorf("Apply(%q): %v", f.name, err)
			continue
		}
		if result == nil {
			t.Errorf("Apply(%q) returned nil", f.name)
			continue
		}
		bounds := result.Bounds()
		if bounds.Dx() != 20 || bounds.Dy() != 20 {
			t.Errorf("Apply(%q) changed dimensions: %dx%d", f.name, bounds.Dx(), bounds.Dy())
		}
	}
}

func TestApplyDefaultParams(t *testing.T) {
	// Filters should work with nil params (using defaults)
	img := createTestImage(10, 10, color.NRGBA{100, 100, 100, 255})

	filtersWithDefaults := []string{
		"blur", "sharpen", "brightness", "contrast", "saturation",
		"gamma", "hue", "pixelate", "vignette", "noise", "posterize", "glow", "duotone",
	}

	for _, name := range filtersWithDefaults {
		result, err := Apply(img, name, nil)
		if err != nil {
			t.Errorf("Apply(%q, nil params): %v", name, err)
		}
		if result == nil {
			t.Errorf("Apply(%q, nil params) returned nil", name)
		}
	}
}

func TestGrayscaleProducesGray(t *testing.T) {
	img := createTestImage(10, 10, color.NRGBA{255, 0, 0, 255})
	result, _ := Apply(img, "grayscale", nil)

	// All channels should be equal for grayscale
	r, g, b, _ := result.At(5, 5).RGBA()
	rr, gg, bb := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	if rr != gg || gg != bb {
		t.Errorf("Grayscale not gray: (%d, %d, %d)", rr, gg, bb)
	}
}

func TestInvertColors(t *testing.T) {
	img := createTestImage(10, 10, color.NRGBA{255, 0, 128, 255})
	result, _ := Apply(img, "invert", nil)

	r, g, b, _ := result.At(5, 5).RGBA()
	rr, gg, bb := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	if rr != 0 || gg != 255 || bb != 127 {
		t.Errorf("Invert: got (%d, %d, %d), want (0, 255, 127)", rr, gg, bb)
	}
}

func TestPixelateMinSize(t *testing.T) {
	img := createTestImage(10, 10, color.NRGBA{128, 128, 128, 255})
	// Size 0 should be clamped to 1
	result, err := Apply(img, "pixelate", map[string]float64{"size": 0})
	if err != nil {
		t.Fatalf("Pixelate size=0: %v", err)
	}
	if result == nil {
		t.Fatal("Pixelate size=0 returned nil")
	}
}

func TestPosterizeMinLevels(t *testing.T) {
	img := createTestImage(10, 10, color.NRGBA{128, 128, 128, 255})
	// Levels < 2 should be clamped to 2
	result, err := Apply(img, "posterize", map[string]float64{"levels": 1})
	if err != nil {
		t.Fatalf("Posterize levels=1: %v", err)
	}
	if result == nil {
		t.Fatal("Posterize levels=1 returned nil")
	}
}

func TestBrightnessAffectsPixels(t *testing.T) {
	img := createTestImage(10, 10, color.NRGBA{128, 128, 128, 255})

	brighter, _ := Apply(img, "brightness", map[string]float64{"value": 50})
	r1, _, _, _ := brighter.At(5, 5).RGBA()

	darker, _ := Apply(img, "brightness", map[string]float64{"value": -50})
	r2, _, _, _ := darker.At(5, 5).RGBA()

	if uint8(r1>>8) <= uint8(r2>>8) {
		t.Errorf("Brighter (%d) should be > darker (%d)", uint8(r1>>8), uint8(r2>>8))
	}
}

func TestSepiaHasWarmTone(t *testing.T) {
	img := createTestImage(10, 10, color.NRGBA{128, 128, 128, 255})
	result, _ := Apply(img, "sepia", nil)

	r, g, b, _ := result.At(5, 5).RGBA()
	rr, gg, bb := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	// Sepia should make R > G > B
	if rr <= gg || gg <= bb {
		t.Errorf("Sepia tone: got (%d, %d, %d), expected R > G > B", rr, gg, bb)
	}
}

// --- getParam ---

func TestGetParam(t *testing.T) {
	params := map[string]float64{"radius": 5.0}
	if got := getParam(params, "radius", 1.0); got != 5.0 {
		t.Errorf("getParam existing: got %f, want 5.0", got)
	}
	if got := getParam(params, "missing", 99.0); got != 99.0 {
		t.Errorf("getParam missing: got %f, want 99.0", got)
	}
	if got := getParam(nil, "anything", 42.0); got != 42.0 {
		t.Errorf("getParam nil map: got %f, want 42.0", got)
	}
}

// --- HSL round-trip ---

func TestHSLRoundTrip(t *testing.T) {
	tests := []struct{ r, g, b float64 }{
		{1, 0, 0},     // red
		{0, 1, 0},     // green
		{0, 0, 1},     // blue
		{1, 1, 1},     // white
		{0, 0, 0},     // black
		{0.5, 0.5, 0.5}, // gray
	}

	for _, tt := range tests {
		h, s, l := rgbToHSL(tt.r, tt.g, tt.b)
		rr, gg, bb := hslToRGB(h, s, l)

		if diff(rr, tt.r) > 0.01 || diff(gg, tt.g) > 0.01 || diff(bb, tt.b) > 0.01 {
			t.Errorf("HSL round-trip (%.1f,%.1f,%.1f) → (%.3f,%.3f,%.3f)",
				tt.r, tt.g, tt.b, rr, gg, bb)
		}
	}
}

func diff(a, b float64) float64 {
	d := a - b
	if d < 0 {
		return -d
	}
	return d
}
