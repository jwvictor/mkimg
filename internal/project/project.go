package project

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const (
	FileExtension = "_mkimg.json"
	Version       = "1"
)

// Canvas defines the output image dimensions and background.
type Canvas struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Background string `json:"background"` // hex color e.g. "#1a1a2e"
}

// Filter represents an effect applied to a layer.
type Filter struct {
	Type   string             `json:"type"`
	Params map[string]float64 `json:"params,omitempty"`
}

// Shadow defines a drop shadow effect.
type Shadow struct {
	Color   string  `json:"color"`
	OffsetX float64 `json:"offset_x"`
	OffsetY float64 `json:"offset_y"`
	Blur    float64 `json:"blur"`
}

// Stroke defines an outline on shapes or text.
type Stroke struct {
	Color string  `json:"color"`
	Width float64 `json:"width"`
}

// GradientStop defines a color stop in a gradient.
type GradientStop struct {
	Color    string  `json:"color"`
	Position float64 `json:"position"` // 0.0 to 1.0
}

// Layer represents a single compositable layer in the project.
type Layer struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"` // solid, image, text, shape, gradient, ai, icon
	Visible bool    `json:"visible"`
	Opacity float64 `json:"opacity"` // 0.0 to 1.0

	// Position & size (optional, defaults to canvas size for full-bleed layers)
	X      float64 `json:"x,omitempty"`
	Y      float64 `json:"y,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`

	// Rotation in degrees
	Rotation float64 `json:"rotation,omitempty"`

	// Type-specific properties

	// Solid
	Color string `json:"color,omitempty"`

	// Image / AI
	Source      string `json:"source,omitempty"`       // file path
	Fit         string `json:"fit,omitempty"`           // cover, contain, fill, none
	AIPrompt    string `json:"ai_prompt,omitempty"`     // for AI-generated layers
	AspectRatio string `json:"aspect_ratio,omitempty"`  // for AI generation

	// Crop (applied to source image before fit/resize)
	CropX      float64 `json:"crop_x,omitempty"`
	CropY      float64 `json:"crop_y,omitempty"`
	CropWidth  float64 `json:"crop_width,omitempty"`
	CropHeight float64 `json:"crop_height,omitempty"`

	// Text
	Content    string  `json:"content,omitempty"`
	Font       string  `json:"font,omitempty"`
	FontSize   float64 `json:"font_size,omitempty"`
	FontWeight string  `json:"font_weight,omitempty"` // regular, bold, italic, etc.
	Align      string  `json:"align,omitempty"`        // left, center, right
	VAlign     string  `json:"valign,omitempty"`       // top, middle, bottom
	LineHeight float64 `json:"line_height,omitempty"`  // multiplier
	MaxWidth   float64 `json:"max_width,omitempty"`    // for text wrapping

	// Shape
	Shape       string  `json:"shape,omitempty"`        // rect, circle, ellipse, line, polygon
	Fill        string  `json:"fill,omitempty"`
	Radius      float64 `json:"radius,omitempty"`       // corner radius for rect, radius for circle
	StrokeStyle *Stroke `json:"stroke,omitempty"`

	// Gradient
	GradientType  string         `json:"gradient_type,omitempty"`  // linear, radial, conic
	GradientStops []GradientStop `json:"gradient_stops,omitempty"`
	GradientAngle float64        `json:"gradient_angle,omitempty"` // degrees, for linear

	// Icon
	IconName       string `json:"icon_name,omitempty"`
	IconCollection string `json:"icon_collection,omitempty"` // material, fontawesome

	// Effects
	Filters []Filter `json:"filters,omitempty"`
	ShadowEffect  *Shadow `json:"shadow,omitempty"`
}

// Project is the top-level structure for a mkimg project file.
type Project struct {
	Version   string    `json:"version"`
	Name      string    `json:"name"`
	Canvas    Canvas    `json:"canvas"`
	Layers    []Layer   `json:"layers"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Runtime (not serialized)
	FilePath string `json:"-"`
}

// New creates a new project with the given parameters.
func New(name string, width, height int, background string) *Project {
	now := time.Now()
	return &Project{
		Version: Version,
		Name:    name,
		Canvas: Canvas{
			Width:      width,
			Height:     height,
			Background: background,
		},
		Layers:    []Layer{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ProjectFile returns the expected filename for a project.
func ProjectFile(name string) string {
	return name + FileExtension
}

// Save writes the project to disk.
func (p *Project) Save() error {
	if p.FilePath == "" {
		p.FilePath = ProjectFile(p.Name)
	}
	p.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal project: %w", err)
	}
	return os.WriteFile(p.FilePath, data, 0644)
}

// Load reads a project from disk.
func Load(path string) (*Project, error) {
	// If path doesn't end with our extension, try appending it
	if filepath.Ext(path) != ".json" {
		candidate := path + FileExtension
		if _, err := os.Stat(candidate); err == nil {
			path = candidate
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read project file: %w", err)
	}

	var p Project
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse project file: %w", err)
	}
	p.FilePath = path
	return &p, nil
}

// FindProject looks for a _mkimg.json file in the current directory.
func FindProject() (*Project, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			name := e.Name()
			if len(name) > len(FileExtension) && name[len(name)-len(FileExtension):] == FileExtension {
				return Load(name)
			}
		}
	}
	return nil, fmt.Errorf("no _mkimg.json project file found in current directory")
}

// GenerateID creates a short random ID for layers.
func GenerateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// AddLayer appends a layer to the project and returns the assigned ID.
func (p *Project) AddLayer(l Layer) string {
	if l.ID == "" {
		l.ID = GenerateID()
	}
	if l.Opacity == 0 {
		l.Opacity = 1.0
	}
	if !l.Visible {
		l.Visible = true
	}
	p.Layers = append(p.Layers, l)
	return l.ID
}

// RemoveLayer removes a layer by ID.
func (p *Project) RemoveLayer(id string) error {
	for i, l := range p.Layers {
		if l.ID == id {
			p.Layers = append(p.Layers[:i], p.Layers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("layer %q not found", id)
}

// GetLayer returns a pointer to a layer by ID or name.
// It checks IDs first, then falls back to matching by name.
func (p *Project) GetLayer(id string) *Layer {
	for i := range p.Layers {
		if p.Layers[i].ID == id {
			return &p.Layers[i]
		}
	}
	// Fallback: match by name
	for i := range p.Layers {
		if p.Layers[i].Name == id {
			return &p.Layers[i]
		}
	}
	return nil
}

// MoveLayer moves a layer to a new position (0-indexed).
func (p *Project) MoveLayer(id string, newPos int) error {
	oldPos := -1
	for i, l := range p.Layers {
		if l.ID == id {
			oldPos = i
			break
		}
	}
	if oldPos == -1 {
		return fmt.Errorf("layer %q not found", id)
	}
	if newPos < 0 || newPos >= len(p.Layers) {
		return fmt.Errorf("position %d out of range (0-%d)", newPos, len(p.Layers)-1)
	}

	layer := p.Layers[oldPos]
	p.Layers = append(p.Layers[:oldPos], p.Layers[oldPos+1:]...)

	newLayers := make([]Layer, 0, len(p.Layers)+1)
	newLayers = append(newLayers, p.Layers[:newPos]...)
	newLayers = append(newLayers, layer)
	newLayers = append(newLayers, p.Layers[newPos:]...)
	p.Layers = newLayers

	return nil
}
