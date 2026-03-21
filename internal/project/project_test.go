package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- Round-trip serialization ---

func TestNewProjectRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := New("test-project", 800, 600, "#ff0000")
	p.FilePath = filepath.Join(dir, "test-project.mkimg.json")

	// Add layers of every type with various fields populated
	p.AddLayer(Layer{Type: "solid", Color: "#00ff00", Name: "bg"})
	p.AddLayer(Layer{
		Type: "text", Content: "Hello World", Font: "Arial",
		FontSize: 32, FontWeight: "bold", Align: "center",
		MaxWidth: 400, LineHeight: 1.5, Color: "#ffffff",
		X: 10, Y: 20, Opacity: 0.8, Rotation: 15,
	})
	p.AddLayer(Layer{
		Type: "shape", Shape: "rect", Fill: "#0000ff",
		Width: 200, Height: 100, Radius: 10,
		StrokeStyle: &Stroke{Color: "#ff0000", Width: 2},
	})
	p.AddLayer(Layer{
		Type: "gradient", GradientType: "linear",
		GradientStops: []GradientStop{
			{Color: "#000000", Position: 0},
			{Color: "#ffffff", Position: 0.5},
			{Color: "#ff0000", Position: 1},
		},
		GradientAngle: 45,
	})
	p.AddLayer(Layer{
		Type: "image", Source: "/tmp/test.png", Fit: "cover",
		Width: 300, Height: 200,
	})
	p.AddLayer(Layer{
		Type: "ai", AIPrompt: "a cat", AspectRatio: "16:9",
		Source: "/tmp/ai.png",
	})
	p.AddLayer(Layer{
		Type: "icon", IconName: "home", IconCollection: "material",
		FontSize: 48, Color: "#ffffff",
	})
	// Layer with filters and shadow
	p.AddLayer(Layer{
		Type: "solid", Color: "#333333",
		Filters: []Filter{
			{Type: "blur", Params: map[string]float64{"radius": 5}},
			{Type: "brightness", Params: map[string]float64{"value": 20}},
		},
		ShadowEffect: &Shadow{Color: "#000000", OffsetX: 5, OffsetY: 5, Blur: 10},
	})

	if err := p.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(p.FilePath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Structural checks
	if loaded.Name != p.Name {
		t.Errorf("Name: got %q, want %q", loaded.Name, p.Name)
	}
	if loaded.Version != Version {
		t.Errorf("Version: got %q, want %q", loaded.Version, Version)
	}
	if loaded.Canvas.Width != 800 || loaded.Canvas.Height != 600 {
		t.Errorf("Canvas: got %dx%d, want 800x600", loaded.Canvas.Width, loaded.Canvas.Height)
	}
	if loaded.Canvas.Background != "#ff0000" {
		t.Errorf("Background: got %q, want %q", loaded.Canvas.Background, "#ff0000")
	}
	if len(loaded.Layers) != len(p.Layers) {
		t.Fatalf("Layers count: got %d, want %d", len(loaded.Layers), len(p.Layers))
	}

	// Verify every layer round-tripped correctly
	for i, orig := range p.Layers {
		got := loaded.Layers[i]
		if got.ID != orig.ID {
			t.Errorf("Layer[%d] ID: got %q, want %q", i, got.ID, orig.ID)
		}
		if got.Type != orig.Type {
			t.Errorf("Layer[%d] Type: got %q, want %q", i, got.Type, orig.Type)
		}
		if got.Opacity != orig.Opacity {
			t.Errorf("Layer[%d] Opacity: got %f, want %f", i, got.Opacity, orig.Opacity)
		}
		if got.Visible != orig.Visible {
			t.Errorf("Layer[%d] Visible: got %v, want %v", i, got.Visible, orig.Visible)
		}
	}

	// Spot-check complex fields
	textLayer := loaded.Layers[1]
	if textLayer.Content != "Hello World" || textLayer.FontWeight != "bold" || textLayer.Align != "center" {
		t.Errorf("Text layer fields mismatch: %+v", textLayer)
	}
	if textLayer.Rotation != 15 {
		t.Errorf("Text layer rotation: got %f, want 15", textLayer.Rotation)
	}

	shapeLayer := loaded.Layers[2]
	if shapeLayer.StrokeStyle == nil {
		t.Fatal("Shape layer stroke lost after round-trip")
	}
	if shapeLayer.StrokeStyle.Color != "#ff0000" || shapeLayer.StrokeStyle.Width != 2 {
		t.Errorf("Shape stroke mismatch: %+v", shapeLayer.StrokeStyle)
	}

	gradLayer := loaded.Layers[3]
	if len(gradLayer.GradientStops) != 3 {
		t.Errorf("Gradient stops: got %d, want 3", len(gradLayer.GradientStops))
	}

	filterLayer := loaded.Layers[7]
	if len(filterLayer.Filters) != 2 {
		t.Errorf("Filters count: got %d, want 2", len(filterLayer.Filters))
	}
	if filterLayer.ShadowEffect == nil {
		t.Fatal("Shadow lost after round-trip")
	}
	if filterLayer.ShadowEffect.Blur != 10 {
		t.Errorf("Shadow blur: got %f, want 10", filterLayer.ShadowEffect.Blur)
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.mkimg.json")
	os.WriteFile(path, []byte(`{not valid json`), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error loading malformed JSON")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/project.mkimg.json")
	if err == nil {
		t.Fatal("expected error loading nonexistent file")
	}
}

func TestLoadExtraFieldsIgnored(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "extra.mkimg.json")
	data := `{
		"version": "1",
		"name": "extra-test",
		"canvas": {"width": 100, "height": 100, "background": "#000"},
		"layers": [],
		"created_at": "2025-01-01T00:00:00Z",
		"updated_at": "2025-01-01T00:00:00Z",
		"unknown_field": "should be ignored"
	}`
	os.WriteFile(path, []byte(data), 0644)

	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load with extra fields: %v", err)
	}
	if p.Name != "extra-test" {
		t.Errorf("Name: got %q, want %q", p.Name, "extra-test")
	}
}

func TestLoadMissingLayerFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sparse.mkimg.json")
	data := `{
		"version": "1",
		"name": "sparse",
		"canvas": {"width": 100, "height": 100, "background": "#000"},
		"layers": [{"id": "abc123", "type": "solid"}],
		"created_at": "2025-01-01T00:00:00Z",
		"updated_at": "2025-01-01T00:00:00Z"
	}`
	os.WriteFile(path, []byte(data), 0644)

	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load sparse layer: %v", err)
	}
	if len(p.Layers) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(p.Layers))
	}
	l := p.Layers[0]
	// Verify zero values for omitted fields
	if l.Color != "" || l.Opacity != 0 || l.Visible != false {
		t.Errorf("Expected zero values for omitted fields, got: color=%q opacity=%f visible=%v",
			l.Color, l.Opacity, l.Visible)
	}
}

func TestSaveProducesValidJSON(t *testing.T) {
	dir := t.TempDir()
	p := New("json-test", 100, 100, "#abc")
	p.FilePath = filepath.Join(dir, "json-test.mkimg.json")
	p.AddLayer(Layer{Type: "solid", Color: "#fff"})

	if err := p.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, _ := os.ReadFile(p.FilePath)
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Saved file is not valid JSON: %v", err)
	}
	// Check version field exists
	if v, ok := raw["version"]; !ok || v != "1" {
		t.Errorf("version field: got %v", v)
	}
}

// --- Layer CRUD ---

func TestAddLayerDefaults(t *testing.T) {
	p := New("test", 100, 100, "#000")

	id := p.AddLayer(Layer{Type: "solid", Color: "#fff"})

	if len(id) != 6 {
		t.Errorf("expected 6-char ID, got %q", id)
	}

	l := p.GetLayer(id)
	if l == nil {
		t.Fatal("GetLayer returned nil")
	}
	if l.Opacity != 1.0 {
		t.Errorf("default Opacity: got %f, want 1.0", l.Opacity)
	}
	if !l.Visible {
		t.Error("default Visible: got false, want true")
	}
}

func TestAddLayerPreservesExplicitID(t *testing.T) {
	p := New("test", 100, 100, "#000")
	id := p.AddLayer(Layer{ID: "custom", Type: "solid"})
	if id != "custom" {
		t.Errorf("expected custom ID, got %q", id)
	}
}

func TestRemoveLayer(t *testing.T) {
	p := New("test", 100, 100, "#000")
	id := p.AddLayer(Layer{Type: "solid"})
	p.AddLayer(Layer{Type: "solid"})

	if err := p.RemoveLayer(id); err != nil {
		t.Fatalf("RemoveLayer: %v", err)
	}
	if len(p.Layers) != 1 {
		t.Errorf("expected 1 layer after remove, got %d", len(p.Layers))
	}
	if p.GetLayer(id) != nil {
		t.Error("removed layer still found")
	}
}

func TestRemoveLayerNotFound(t *testing.T) {
	p := New("test", 100, 100, "#000")
	if err := p.RemoveLayer("nonexistent"); err == nil {
		t.Error("expected error removing nonexistent layer")
	}
}

func TestGetLayerNotFound(t *testing.T) {
	p := New("test", 100, 100, "#000")
	if p.GetLayer("nope") != nil {
		t.Error("expected nil for nonexistent layer")
	}
}

// --- MoveLayer ---

func TestMoveLayerForward(t *testing.T) {
	p := New("test", 100, 100, "#000")
	idA := p.AddLayer(Layer{ID: "a", Type: "solid"})
	p.AddLayer(Layer{ID: "b", Type: "solid"})
	p.AddLayer(Layer{ID: "c", Type: "solid"})

	// Move "a" from position 0 to position 2
	if err := p.MoveLayer(idA, 2); err != nil {
		t.Fatalf("MoveLayer: %v", err)
	}

	expected := []string{"b", "c", "a"}
	for i, want := range expected {
		if p.Layers[i].ID != want {
			t.Errorf("position %d: got %q, want %q", i, p.Layers[i].ID, want)
		}
	}
}

func TestMoveLayerBackward(t *testing.T) {
	p := New("test", 100, 100, "#000")
	p.AddLayer(Layer{ID: "a", Type: "solid"})
	p.AddLayer(Layer{ID: "b", Type: "solid"})
	p.AddLayer(Layer{ID: "c", Type: "solid"})

	// Move "c" from position 2 to position 0
	if err := p.MoveLayer("c", 0); err != nil {
		t.Fatalf("MoveLayer: %v", err)
	}

	expected := []string{"c", "a", "b"}
	for i, want := range expected {
		if p.Layers[i].ID != want {
			t.Errorf("position %d: got %q, want %q", i, p.Layers[i].ID, want)
		}
	}
}

func TestMoveLayerSamePosition(t *testing.T) {
	p := New("test", 100, 100, "#000")
	p.AddLayer(Layer{ID: "a", Type: "solid"})
	p.AddLayer(Layer{ID: "b", Type: "solid"})
	p.AddLayer(Layer{ID: "c", Type: "solid"})

	if err := p.MoveLayer("b", 1); err != nil {
		t.Fatalf("MoveLayer same position: %v", err)
	}

	expected := []string{"a", "b", "c"}
	for i, want := range expected {
		if p.Layers[i].ID != want {
			t.Errorf("position %d: got %q, want %q", i, p.Layers[i].ID, want)
		}
	}
}

func TestMoveLayerToMiddle(t *testing.T) {
	p := New("test", 100, 100, "#000")
	p.AddLayer(Layer{ID: "a", Type: "solid"})
	p.AddLayer(Layer{ID: "b", Type: "solid"})
	p.AddLayer(Layer{ID: "c", Type: "solid"})
	p.AddLayer(Layer{ID: "d", Type: "solid"})

	// Move "a" from position 0 to position 2
	if err := p.MoveLayer("a", 2); err != nil {
		t.Fatalf("MoveLayer: %v", err)
	}

	expected := []string{"b", "c", "a", "d"}
	for i, want := range expected {
		if p.Layers[i].ID != want {
			t.Errorf("position %d: got %q, want %q", i, p.Layers[i].ID, want)
		}
	}
}

func TestMoveLayerNotFound(t *testing.T) {
	p := New("test", 100, 100, "#000")
	p.AddLayer(Layer{Type: "solid"})
	if err := p.MoveLayer("nonexistent", 0); err == nil {
		t.Error("expected error moving nonexistent layer")
	}
}

func TestMoveLayerOutOfBounds(t *testing.T) {
	p := New("test", 100, 100, "#000")
	id := p.AddLayer(Layer{Type: "solid"})

	if err := p.MoveLayer(id, -1); err == nil {
		t.Error("expected error for negative position")
	}
	if err := p.MoveLayer(id, 5); err == nil {
		t.Error("expected error for position beyond length")
	}
}

// --- FindProject ---

func TestFindProjectInDirectory(t *testing.T) {
	dir := t.TempDir()

	p := New("findme", 100, 100, "#000")
	p.FilePath = filepath.Join(dir, ProjectFile("findme"))
	if err := p.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify the file actually exists
	entries, _ := os.ReadDir(dir)
	found := false
	for _, e := range entries {
		if e.Name() == ProjectFile("findme") {
			found = true
		}
	}
	if !found {
		t.Fatalf("project file not created in %s", dir)
	}

	// Change to temp dir to test FindProject
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(origDir)

	proj, err := FindProject()
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if proj.Name != "findme" {
		t.Errorf("found wrong project: got %q, want %q", proj.Name, "findme")
	}
}

func TestFindProjectNoProject(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	_, err := FindProject()
	if err == nil {
		t.Error("expected error when no project exists")
	}
}

// --- GenerateID ---

func TestGenerateIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateID()
		if len(id) != 6 {
			t.Fatalf("ID length: got %d, want 6", len(id))
		}
		if seen[id] {
			t.Fatalf("duplicate ID generated: %q", id)
		}
		seen[id] = true
	}
}

// --- Load path resolution ---

func TestLoadWithoutExtension(t *testing.T) {
	dir := t.TempDir()
	p := New("myproj", 100, 100, "#000")
	p.FilePath = filepath.Join(dir, ProjectFile("myproj"))
	if err := p.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Load using just the name (without .mkimg.json extension)
	loaded, err := Load(filepath.Join(dir, "myproj"))
	if err != nil {
		t.Fatalf("Load without extension: %v", err)
	}
	if loaded.Name != "myproj" {
		t.Errorf("Name: got %q, want %q", loaded.Name, "myproj")
	}
}
