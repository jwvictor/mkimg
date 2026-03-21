package engine

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"mkimg/internal/project"
)

// --- parseHexColor ---

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		input string
		want  color.NRGBA
	}{
		{"#ff0000", color.NRGBA{255, 0, 0, 255}},
		{"#00ff00", color.NRGBA{0, 255, 0, 255}},
		{"#0000ff", color.NRGBA{0, 0, 255, 255}},
		{"#000000", color.NRGBA{0, 0, 0, 255}},
		{"#ffffff", color.NRGBA{255, 255, 255, 255}},
		{"ff0000", color.NRGBA{255, 0, 0, 255}},          // no hash
		{"#f00", color.NRGBA{255, 0, 0, 255}},             // 3-char
		{"#0f0", color.NRGBA{0, 255, 0, 255}},             // 3-char
		{"#ff000080", color.NRGBA{255, 0, 0, 128}},        // 8-char with alpha
		{"#f008", color.NRGBA{255, 0, 0, 136}},            // 4-char with alpha
		{"#1a1a2e", color.NRGBA{26, 26, 46, 255}},         // real-world color
	}

	for _, tt := range tests {
		got := parseHexColor(tt.input)
		gotNRGBA, ok := got.(color.NRGBA)
		if !ok {
			t.Errorf("parseHexColor(%q): not NRGBA, got %T", tt.input, got)
			continue
		}
		if gotNRGBA != tt.want {
			t.Errorf("parseHexColor(%q) = %v, want %v", tt.input, gotNRGBA, tt.want)
		}
	}
}

func TestParseHexColorEmpty(t *testing.T) {
	got := parseHexColor("")
	if got != color.Transparent {
		t.Errorf("parseHexColor(\"\") = %v, want Transparent", got)
	}
}

func TestParseHexColorInvalid(t *testing.T) {
	// Invalid hex should return black (current behavior)
	invalids := []string{"#gg", "#12345", "xyz", "#1"}
	for _, s := range invalids {
		got := parseHexColor(s)
		r, g, b, _ := got.RGBA()
		if r != 0 || g != 0 || b != 0 {
			t.Errorf("parseHexColor(%q) = %v, expected black", s, got)
		}
	}
}

// --- intOrDefault ---

func TestIntOrDefault(t *testing.T) {
	if got := intOrDefault(50, 100); got != 50 {
		t.Errorf("intOrDefault(50, 100) = %d, want 50", got)
	}
	if got := intOrDefault(0, 100); got != 100 {
		t.Errorf("intOrDefault(0, 100) = %d, want 100", got)
	}
	if got := intOrDefault(-5, 100); got != 100 {
		t.Errorf("intOrDefault(-5, 100) = %d, want 100", got)
	}
}

// --- axToAlign ---

func TestAxToAlign(t *testing.T) {
	if got := axToAlign(0.0); got != 0 {
		t.Errorf("axToAlign(0.0) = %d, want 0 (left)", got)
	}
	if got := axToAlign(0.5); got != 1 {
		t.Errorf("axToAlign(0.5) = %d, want 1 (center)", got)
	}
	if got := axToAlign(1.0); got != 2 {
		t.Errorf("axToAlign(1.0) = %d, want 2 (right)", got)
	}
}

// --- applyOpacity ---

func TestApplyOpacity(t *testing.T) {
	// Create a 2x2 solid red image
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetNRGBA(x, y, color.NRGBA{255, 0, 0, 255})
		}
	}

	result := applyOpacity(img, 0.5)
	nrgba := result.(*image.NRGBA)
	c := nrgba.NRGBAAt(0, 0)

	if c.R != 255 {
		t.Errorf("Red channel changed: got %d, want 255", c.R)
	}
	// Alpha should be ~127 (255 * 0.5)
	if c.A < 126 || c.A > 128 {
		t.Errorf("Alpha: got %d, want ~127", c.A)
	}
}

func TestApplyOpacityFull(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{100, 150, 200, 255})

	result := applyOpacity(img, 1.0)
	nrgba := result.(*image.NRGBA)
	c := nrgba.NRGBAAt(0, 0)

	if c.A != 255 {
		t.Errorf("Full opacity should preserve alpha: got %d", c.A)
	}
}

func TestApplyOpacityZero(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{100, 150, 200, 255})

	result := applyOpacity(img, 0.0)
	nrgba := result.(*image.NRGBA)
	c := nrgba.NRGBAAt(0, 0)

	if c.A != 0 {
		t.Errorf("Zero opacity should zero alpha: got %d", c.A)
	}
}

// --- lerpColor ---

func TestLerpColor(t *testing.T) {
	c1 := color.NRGBA{0, 0, 0, 255}
	c2 := color.NRGBA{255, 255, 255, 255}

	mid := lerpColor(c1, c2, 0.5)
	r, g, b, _ := mid.RGBA()
	// At t=0.5, should be roughly 127-128 for each channel
	rr := uint8(r >> 8)
	if rr < 126 || rr > 129 {
		t.Errorf("lerpColor midpoint R: got %d, want ~127", rr)
	}

	gg := uint8(g >> 8)
	if gg < 126 || gg > 129 {
		t.Errorf("lerpColor midpoint G: got %d, want ~127", gg)
	}

	bb := uint8(b >> 8)
	if bb < 126 || bb > 129 {
		t.Errorf("lerpColor midpoint B: got %d, want ~127", bb)
	}
}

func TestLerpColorEndpoints(t *testing.T) {
	c1 := color.NRGBA{255, 0, 0, 255}
	c2 := color.NRGBA{0, 0, 255, 255}

	// t=0 should give c1
	start := lerpColor(c1, c2, 0)
	r, _, _, _ := start.RGBA()
	if uint8(r>>8) != 255 {
		t.Errorf("lerpColor(t=0) R: got %d, want 255", uint8(r>>8))
	}

	// t=1 should give c2
	end := lerpColor(c1, c2, 1.0)
	_, _, b, _ := end.RGBA()
	if uint8(b>>8) != 255 {
		t.Errorf("lerpColor(t=1) B: got %d, want 255", uint8(b>>8))
	}
}

// --- interpolateStops ---

func TestInterpolateStopsEdges(t *testing.T) {
	stops := []project.GradientStop{
		{Color: "#ff0000", Position: 0.0},
		{Color: "#0000ff", Position: 1.0},
	}

	// Before first stop
	c := interpolateStops(stops, -0.5)
	r, _, _, _ := c.RGBA()
	if uint8(r>>8) != 255 {
		t.Errorf("Before first stop: R should be 255, got %d", uint8(r>>8))
	}

	// After last stop
	c = interpolateStops(stops, 1.5)
	_, _, b, _ := c.RGBA()
	if uint8(b>>8) != 255 {
		t.Errorf("After last stop: B should be 255, got %d", uint8(b>>8))
	}
}

// --- Rendering integration tests ---

func TestRenderSolidLayer(t *testing.T) {
	p := project.New("test", 100, 100, "#000000")
	p.AddLayer(project.Layer{
		Type:  "solid",
		Color: "#ff0000",
	})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Output size: %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}

	// Center pixel should be red
	r, g, b, _ := img.At(50, 50).RGBA()
	if uint8(r>>8) != 255 || uint8(g>>8) != 0 || uint8(b>>8) != 0 {
		t.Errorf("Center pixel: got (%d,%d,%d), want (255,0,0)",
			uint8(r>>8), uint8(g>>8), uint8(b>>8))
	}
}

func TestRenderEmptyProject(t *testing.T) {
	p := project.New("test", 50, 50, "#ffffff")

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Should be all white
	r, g, b, _ := img.At(25, 25).RGBA()
	if uint8(r>>8) != 255 || uint8(g>>8) != 255 || uint8(b>>8) != 255 {
		t.Errorf("Empty project pixel: got (%d,%d,%d), want (255,255,255)",
			uint8(r>>8), uint8(g>>8), uint8(b>>8))
	}
}

func TestRenderInvisibleLayerSkipped(t *testing.T) {
	p := project.New("test", 50, 50, "#ffffff")
	p.Layers = append(p.Layers, project.Layer{
		ID: "hidden", Type: "solid", Color: "#ff0000",
		Visible: false, Opacity: 1.0,
	})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Should still be white — red layer is invisible
	r, g, b, _ := img.At(25, 25).RGBA()
	if uint8(r>>8) != 255 || uint8(g>>8) != 255 || uint8(b>>8) != 255 {
		t.Errorf("Invisible layer leaked: got (%d,%d,%d)", uint8(r>>8), uint8(g>>8), uint8(b>>8))
	}
}

func TestRenderLayerOrdering(t *testing.T) {
	p := project.New("test", 50, 50, "#000000")
	// First layer: red (will be covered)
	p.AddLayer(project.Layer{Type: "solid", Color: "#ff0000"})
	// Second layer: blue (on top)
	p.AddLayer(project.Layer{Type: "solid", Color: "#0000ff"})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Should be blue (top layer wins)
	r, g, b, _ := img.At(25, 25).RGBA()
	if uint8(r>>8) != 0 || uint8(g>>8) != 0 || uint8(b>>8) != 255 {
		t.Errorf("Layer ordering: got (%d,%d,%d), want (0,0,255)",
			uint8(r>>8), uint8(g>>8), uint8(b>>8))
	}
}

func TestRenderOpacity(t *testing.T) {
	p := project.New("test", 50, 50, "#000000")
	// Half-opacity white over black should give ~127 gray
	p.Layers = append(p.Layers, project.Layer{
		ID: "semi", Type: "solid", Color: "#ffffff",
		Visible: true, Opacity: 0.5,
	})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	r, g, b, _ := img.At(25, 25).RGBA()
	rr := uint8(r >> 8)
	// Should be roughly midway (alpha blending over black)
	if rr < 100 || rr > 155 {
		t.Errorf("Opacity blending: got R=%d (expected ~127)", rr)
	}
	gg := uint8(g >> 8)
	bb := uint8(b >> 8)
	if gg < 100 || gg > 155 || bb < 100 || bb > 155 {
		t.Errorf("Opacity blending: got (%d,%d,%d)", rr, gg, bb)
	}
}

func TestRenderShapeRect(t *testing.T) {
	p := project.New("test", 100, 100, "#000000")
	p.AddLayer(project.Layer{
		Type: "shape", Shape: "rect", Fill: "#00ff00",
		Width: 50, Height: 50, X: 25, Y: 25,
	})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Center of the rect (50,50) should be green
	r, g, b, _ := img.At(50, 50).RGBA()
	if uint8(g>>8) != 255 {
		t.Errorf("Rect center: got G=%d, want 255", uint8(g>>8))
	}

	// Corner (0,0) should be black (outside rect)
	r, g, b, _ = img.At(0, 0).RGBA()
	if uint8(r>>8) != 0 || uint8(g>>8) != 0 || uint8(b>>8) != 0 {
		t.Errorf("Outside rect: got (%d,%d,%d), want (0,0,0)",
			uint8(r>>8), uint8(g>>8), uint8(b>>8))
	}
}

func TestRenderShapeCircle(t *testing.T) {
	p := project.New("test", 100, 100, "#000000")
	p.AddLayer(project.Layer{
		Type: "shape", Shape: "circle", Fill: "#ff0000",
		Width: 80, Height: 80, X: 10, Y: 10,
	})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Center of circle should be red
	r, _, _, _ := img.At(50, 50).RGBA()
	if uint8(r>>8) < 200 {
		t.Errorf("Circle center R: got %d, want ~255", uint8(r>>8))
	}
}

func TestRenderGradientLinear(t *testing.T) {
	p := project.New("test", 100, 100, "#000000")
	p.AddLayer(project.Layer{
		Type:         "gradient",
		GradientType: "linear",
		GradientStops: []project.GradientStop{
			{Color: "#000000", Position: 0},
			{Color: "#ffffff", Position: 1},
		},
		GradientAngle: 0,
	})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Just verify it rendered without error and has correct dimensions
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Gradient output: %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}
}

func TestRenderGradientDefaultStops(t *testing.T) {
	p := project.New("test", 50, 50, "#000000")
	// Gradient with no stops — should default to black→white
	p.AddLayer(project.Layer{
		Type:         "gradient",
		GradientType: "linear",
	})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
}

func TestRenderImageLayer(t *testing.T) {
	dir := t.TempDir()

	// Create a test PNG
	testImg := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			testImg.SetNRGBA(x, y, color.NRGBA{0, 255, 0, 255})
		}
	}
	imgPath := filepath.Join(dir, "test.png")
	f, _ := os.Create(imgPath)
	png.Encode(f, testImg)
	f.Close()

	p := project.New("test", 100, 100, "#000000")
	p.AddLayer(project.Layer{
		Type: "image", Source: imgPath,
		Width: 50, Height: 50, Fit: "fill",
		X: 25, Y: 25,
	})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Center should be greenish (the resized test image)
	_, g, _, _ := img.At(50, 50).RGBA()
	if uint8(g>>8) < 200 {
		t.Errorf("Image layer center G: got %d, want ~255", uint8(g>>8))
	}
}

func TestRenderImageFitModes(t *testing.T) {
	dir := t.TempDir()

	// Create a non-square test image (40x20)
	testImg := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			testImg.SetNRGBA(x, y, color.NRGBA{255, 0, 0, 255})
		}
	}
	imgPath := filepath.Join(dir, "wide.png")
	f, _ := os.Create(imgPath)
	png.Encode(f, testImg)
	f.Close()

	fits := []string{"cover", "contain", "fill", "none"}
	for _, fit := range fits {
		p := project.New("test", 100, 100, "#000000")
		p.AddLayer(project.Layer{
			Type: "image", Source: imgPath,
			Width: 60, Height: 60, Fit: fit,
		})

		img, err := Render(p)
		if err != nil {
			t.Errorf("Render with fit=%q: %v", fit, err)
			continue
		}
		if img == nil {
			t.Errorf("Render with fit=%q returned nil", fit)
		}
	}
}

func TestRenderImageMissingSource(t *testing.T) {
	p := project.New("test", 50, 50, "#000")
	p.AddLayer(project.Layer{Type: "image", Source: ""})

	_, err := Render(p)
	if err == nil {
		t.Error("expected error for image layer with no source")
	}
}

func TestRenderImageMissingFile(t *testing.T) {
	p := project.New("test", 50, 50, "#000")
	p.AddLayer(project.Layer{Type: "image", Source: "/nonexistent/file.png"})

	_, err := Render(p)
	if err == nil {
		t.Error("expected error for missing image file")
	}
}

func TestRenderUnknownLayerType(t *testing.T) {
	p := project.New("test", 50, 50, "#000")
	p.Layers = append(p.Layers, project.Layer{
		ID: "bad", Type: "hologram", Visible: true, Opacity: 1.0,
	})

	_, err := Render(p)
	if err == nil {
		t.Error("expected error for unknown layer type")
	}
}

// --- SaveImage ---

func TestSaveImagePNG(t *testing.T) {
	dir := t.TempDir()
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	path := filepath.Join(dir, "out.png")

	if err := SaveImage(img, path); err != nil {
		t.Fatalf("SaveImage PNG: %v", err)
	}

	// Verify it's a valid PNG
	f, _ := os.Open(path)
	defer f.Close()
	_, err := png.Decode(f)
	if err != nil {
		t.Errorf("Saved file is not valid PNG: %v", err)
	}
}

func TestSaveImageJPEG(t *testing.T) {
	dir := t.TempDir()
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	path := filepath.Join(dir, "out.jpg")

	if err := SaveImage(img, path); err != nil {
		t.Fatalf("SaveImage JPEG: %v", err)
	}

	info, _ := os.Stat(path)
	if info.Size() == 0 {
		t.Error("JPEG file is empty")
	}
}

// --- Render with filters ---

func TestRenderLayerWithFilter(t *testing.T) {
	p := project.New("test", 50, 50, "#000000")
	p.Layers = append(p.Layers, project.Layer{
		ID: "filtered", Type: "solid", Color: "#ffffff",
		Visible: true, Opacity: 1.0,
		Filters: []project.Filter{
			{Type: "blur", Params: map[string]float64{"radius": 2}},
		},
	})

	img, err := Render(p)
	if err != nil {
		t.Fatalf("Render with filter: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
}

func TestRenderLayerWithBadFilter(t *testing.T) {
	p := project.New("test", 50, 50, "#000")
	p.Layers = append(p.Layers, project.Layer{
		ID: "bad", Type: "solid", Color: "#fff",
		Visible: true, Opacity: 1.0,
		Filters: []project.Filter{
			{Type: "nonexistent_filter"},
		},
	})

	_, err := Render(p)
	if err == nil {
		t.Error("expected error for unknown filter type")
	}
}
