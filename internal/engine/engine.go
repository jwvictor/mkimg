package engine

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/jwvictor/mkimg/internal/effects"
	"github.com/jwvictor/mkimg/internal/fonts"
	"github.com/jwvictor/mkimg/internal/icons"
	"github.com/jwvictor/mkimg/internal/project"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	_ "golang.org/x/image/webp"
)

// Render composites all visible layers and returns the final image.
func Render(p *project.Project) (image.Image, error) {
	w, h := p.Canvas.Width, p.Canvas.Height

	// Create canvas
	dc := gg.NewContext(w, h)

	// Fill background
	bg := parseHexColor(p.Canvas.Background)
	dc.SetColor(bg)
	dc.Clear()

	// Render each visible layer bottom-to-top
	for i, layer := range p.Layers {
		if !layer.Visible {
			continue
		}

		layerImg, err := renderLayer(dc, &layer, w, h)
		if err != nil {
			return nil, fmt.Errorf("render layer %d (%s): %w", i, layer.ID, err)
		}

		if layerImg != nil {
			// Apply filters to the layer image
			for _, f := range layer.Filters {
				layerImg, err = effects.Apply(layerImg, f.Type, f.Params)
				if err != nil {
					return nil, fmt.Errorf("apply filter %s to layer %s: %w", f.Type, layer.ID, err)
				}
			}

			// Apply opacity
			if layer.Opacity < 1.0 {
				layerImg = applyOpacity(layerImg, layer.Opacity)
			}

			// Apply rotation if specified
			if layer.Rotation != 0 {
				layerImg = imaging.Rotate(layerImg, layer.Rotation, color.Transparent)
			}

			// Apply shadow if specified
			if layer.ShadowEffect != nil {
				drawShadow(dc, layerImg, &layer)
			}

			// Composite onto canvas
			dc.DrawImage(layerImg, int(layer.X), int(layer.Y))
		}
	}

	return dc.Image(), nil
}

// RenderToFile renders the project and saves to a file.
func RenderToFile(p *project.Project, outputPath string) error {
	img, err := Render(p)
	if err != nil {
		return err
	}
	return SaveImage(img, outputPath)
}

// SaveImage writes an image to disk, detecting format from extension.
func SaveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 95})
	case ".png":
		return png.Encode(f, img)
	default:
		return png.Encode(f, img)
	}
}

func renderLayer(dc *gg.Context, layer *project.Layer, canvasW, canvasH int) (image.Image, error) {
	switch layer.Type {
	case "solid":
		return renderSolid(layer, canvasW, canvasH), nil
	case "gradient":
		return renderGradient(layer, canvasW, canvasH), nil
	case "image":
		return renderImage(layer, canvasW, canvasH)
	case "ai":
		return renderImage(layer, canvasW, canvasH) // AI layers reference generated files
	case "text":
		return renderText(layer, canvasW, canvasH)
	case "shape":
		return renderShape(layer, canvasW, canvasH), nil
	case "icon":
		return renderIcon(layer, canvasW, canvasH)
	default:
		return nil, fmt.Errorf("unknown layer type: %s", layer.Type)
	}
}

func renderSolid(layer *project.Layer, canvasW, canvasH int) image.Image {
	w := intOrDefault(layer.Width, float64(canvasW))
	h := intOrDefault(layer.Height, float64(canvasH))
	dc := gg.NewContext(w, h)
	dc.SetColor(parseHexColor(layer.Color))
	dc.Clear()
	return dc.Image()
}

func renderGradient(layer *project.Layer, canvasW, canvasH int) image.Image {
	w := intOrDefault(layer.Width, float64(canvasW))
	h := intOrDefault(layer.Height, float64(canvasH))
	dc := gg.NewContext(w, h)

	stops := layer.GradientStops
	if len(stops) < 2 {
		// Default to a simple two-color gradient
		stops = []project.GradientStop{
			{Color: "#000000", Position: 0},
			{Color: "#ffffff", Position: 1},
		}
	}

	switch layer.GradientType {
	case "radial":
		cx, cy := float64(w)/2, float64(h)/2
		maxR := math.Sqrt(cx*cx + cy*cy)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dx, dy := float64(x)-cx, float64(y)-cy
				dist := math.Sqrt(dx*dx+dy*dy) / maxR
				c := interpolateStops(stops, dist)
				dc.SetColor(c)
				dc.SetPixel(x, y)
			}
		}
	case "conic":
		cx, cy := float64(w)/2, float64(h)/2
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				angle := math.Atan2(float64(y)-cy, float64(x)-cx)
				t := (angle + math.Pi) / (2 * math.Pi)
				c := interpolateStops(stops, t)
				dc.SetColor(c)
				dc.SetPixel(x, y)
			}
		}
	default: // linear
		angleRad := layer.GradientAngle * math.Pi / 180.0
		dx := math.Cos(angleRad)
		dy := math.Sin(angleRad)

		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				// Project point onto gradient line
				nx := float64(x) / float64(w)
				ny := float64(y) / float64(h)
				t := nx*dx + ny*dy
				// Normalize to 0-1
				t = math.Max(0, math.Min(1, (t+1)/2))
				c := interpolateStops(stops, t)
				dc.SetColor(c)
				dc.SetPixel(x, y)
			}
		}
	}

	return dc.Image()
}

func renderImage(layer *project.Layer, canvasW, canvasH int) (image.Image, error) {
	if layer.Source == "" {
		return nil, fmt.Errorf("image layer %s has no source", layer.ID)
	}

	img, err := imaging.Open(layer.Source)
	if err != nil {
		return nil, fmt.Errorf("open image %s: %w", layer.Source, err)
	}

	// Apply crop to source image before any fit/resize
	if layer.CropWidth > 0 && layer.CropHeight > 0 {
		cx := int(layer.CropX)
		cy := int(layer.CropY)
		cw := int(layer.CropWidth)
		ch := int(layer.CropHeight)
		img = imaging.Crop(img, image.Rect(cx, cy, cx+cw, cy+ch))
	}

	// For cover/contain/fill, default to canvas dimensions if not specified.
	// For other modes, default to source image dimensions.
	defaultW, defaultH := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())
	if layer.Fit == "cover" || layer.Fit == "contain" || layer.Fit == "fill" {
		defaultW, defaultH = float64(canvasW), float64(canvasH)
	}
	w := intOrDefault(layer.Width, defaultW)
	h := intOrDefault(layer.Height, defaultH)

	switch layer.Fit {
	case "cover":
		img = imaging.Fill(img, w, h, imaging.Center, imaging.Lanczos)
	case "contain":
		img = imaging.Fit(img, w, h, imaging.Lanczos)
	case "fill":
		img = imaging.Resize(img, w, h, imaging.Lanczos)
	case "none":
		// Keep original size
	default:
		if layer.Width > 0 || layer.Height > 0 {
			rw, rh := 0, 0
			if layer.Width > 0 {
				rw = w
			}
			if layer.Height > 0 {
				rh = h
			}
			img = imaging.Resize(img, rw, rh, imaging.Lanczos)
		}
	}

	return img, nil
}

func renderText(layer *project.Layer, canvasW, canvasH int) (image.Image, error) {
	// Determine text area size
	maxW := intOrDefault(layer.MaxWidth, float64(canvasW)-layer.X)
	fontSize := layer.FontSize
	if fontSize == 0 {
		fontSize = 24
	}

	// Load font
	fontFamily := layer.Font
	if fontFamily == "" {
		fontFamily = "Arial"
	}
	variant := "regular"
	if layer.FontWeight == "bold" {
		variant = "700"
	} else if layer.FontWeight == "italic" {
		variant = "italic"
	} else if layer.FontWeight != "" {
		variant = layer.FontWeight
	}

	face, err := fonts.LoadFace(fontFamily, fontSize, variant)
	if err != nil {
		// Fall back to system default
		face, err = fonts.LoadFace("Helvetica", fontSize, "regular")
		if err != nil {
			return nil, fmt.Errorf("load font: %w", err)
		}
	}

	// Create a context for text measurement and rendering
	dc := gg.NewContext(canvasW, canvasH)
	dc.SetFontFace(face)

	textColor := parseHexColor(layer.Color)
	if layer.Color == "" {
		textColor = color.Black
	}
	dc.SetColor(textColor)

	lineSpacing := layer.LineHeight
	if lineSpacing == 0 {
		lineSpacing = 1.4
	}

	// Determine alignment
	ax := 0.0
	switch layer.Align {
	case "center":
		ax = 0.5
	case "right":
		ax = 1.0
	}

	// Draw text — gg uses baseline Y, so we draw at (0, fontSize) to keep text visible,
	// then the main render loop composites this layer image at (layer.X, layer.Y).
	if layer.MaxWidth > 0 || layer.Width > 0 {
		dc.DrawStringWrapped(layer.Content, 0, 0, ax, 0, float64(maxW), lineSpacing, gg.Align(axToAlign(ax)))
	} else {
		dc.DrawString(layer.Content, 0, fontSize)
	}

	return dc.Image(), nil
}

func renderShape(layer *project.Layer, canvasW, canvasH int) image.Image {
	w := intOrDefault(layer.Width, 100)
	h := intOrDefault(layer.Height, 100)

	// Add padding for stroke
	pad := 0.0
	if layer.StrokeStyle != nil {
		pad = layer.StrokeStyle.Width
	}
	imgW := w + int(pad*2)
	imgH := h + int(pad*2)

	dc := gg.NewContext(imgW, imgH)

	fillColor := parseHexColor(layer.Fill)
	if layer.Fill == "" {
		fillColor = color.Transparent
	}

	switch layer.Shape {
	case "rect":
		if layer.Radius > 0 {
			dc.DrawRoundedRectangle(pad, pad, float64(w), float64(h), layer.Radius)
		} else {
			dc.DrawRectangle(pad, pad, float64(w), float64(h))
		}
	case "circle":
		r := float64(w) / 2
		dc.DrawCircle(float64(imgW)/2, float64(imgH)/2, r)
	case "ellipse":
		dc.DrawEllipse(float64(imgW)/2, float64(imgH)/2, float64(w)/2, float64(h)/2)
	case "line":
		dc.DrawLine(pad, pad, float64(w)+pad, float64(h)+pad)
	default:
		dc.DrawRectangle(pad, pad, float64(w), float64(h))
	}

	if layer.Shape != "line" {
		dc.SetColor(fillColor)
		dc.Fill()

		// Re-draw path for stroke
		if layer.StrokeStyle != nil {
			switch layer.Shape {
			case "rect":
				if layer.Radius > 0 {
					dc.DrawRoundedRectangle(pad, pad, float64(w), float64(h), layer.Radius)
				} else {
					dc.DrawRectangle(pad, pad, float64(w), float64(h))
				}
			case "circle":
				r := float64(w) / 2
				dc.DrawCircle(float64(imgW)/2, float64(imgH)/2, r)
			case "ellipse":
				dc.DrawEllipse(float64(imgW)/2, float64(imgH)/2, float64(w)/2, float64(h)/2)
			}
		}
	}

	if layer.StrokeStyle != nil {
		dc.SetColor(parseHexColor(layer.StrokeStyle.Color))
		dc.SetLineWidth(layer.StrokeStyle.Width)
		dc.Stroke()
	}

	return dc.Image()
}

func renderIcon(layer *project.Layer, canvasW, canvasH int) (image.Image, error) {
	collection := layer.IconCollection
	if collection == "" {
		collection = "material"
	}

	fontSize := layer.FontSize
	if fontSize == 0 {
		fontSize = 48
	}

	var fontPath string
	var cp rune

	switch collection {
	case "material":
		var err error
		cp, err = icons.LookupMaterialIcon(layer.IconName)
		if err != nil {
			return nil, err
		}
		fontPath = icons.MaterialFontPath()
	case "fontawesome":
		var style string
		var err error
		cp, style, err = icons.LookupFontAwesomeIcon(layer.IconName)
		if err != nil {
			return nil, err
		}
		fontPath = icons.FontAwesomeFontPath(style)
	default:
		return nil, fmt.Errorf("unknown icon collection: %s", collection)
	}

	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("icon font not installed — run: mkimg icon install %s", collection)
	}

	dc := gg.NewContext(int(fontSize*2), int(fontSize*2))
	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		return nil, fmt.Errorf("load icon font: %w", err)
	}

	iconColor := parseHexColor(layer.Color)
	if layer.Color == "" {
		iconColor = color.White
	}
	dc.SetColor(iconColor)
	dc.DrawStringAnchored(string(cp), fontSize, fontSize, 0.5, 0.5)

	return dc.Image(), nil
}

// Helper functions

func applyOpacity(img image.Image, opacity float64) image.Image {
	bounds := img.Bounds()
	dst := image.NewNRGBA(bounds)
	draw.Draw(dst, bounds, img, bounds.Min, draw.Src)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := dst.NRGBAAt(x, y)
			c.A = uint8(float64(c.A) * opacity)
			dst.SetNRGBA(x, y, c)
		}
	}
	return dst
}

func drawShadow(dc *gg.Context, layerImg image.Image, layer *project.Layer) {
	s := layer.ShadowEffect
	shadowColor := parseHexColor(s.Color)

	bounds := layerImg.Bounds()
	shadowDC := gg.NewContext(bounds.Dx()+int(s.Blur*4), bounds.Dy()+int(s.Blur*4))

	// Create shadow silhouette
	shadow := image.NewNRGBA(bounds)
	r, g, b, _ := shadowColor.RGBA()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := layerImg.At(x, y).RGBA()
			shadow.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	// Blur the shadow
	blurred := imaging.Blur(shadow, s.Blur)
	shadowDC.DrawImage(blurred, 0, 0)

	dc.DrawImage(shadowDC.Image(), int(layer.X+s.OffsetX), int(layer.Y+s.OffsetY))
}

func interpolateStops(stops []project.GradientStop, t float64) color.Color {
	if t <= stops[0].Position {
		return parseHexColor(stops[0].Color)
	}
	if t >= stops[len(stops)-1].Position {
		return parseHexColor(stops[len(stops)-1].Color)
	}

	for i := 0; i < len(stops)-1; i++ {
		if t >= stops[i].Position && t <= stops[i+1].Position {
			// Interpolate between stops[i] and stops[i+1]
			range_ := stops[i+1].Position - stops[i].Position
			if range_ == 0 {
				return parseHexColor(stops[i].Color)
			}
			localT := (t - stops[i].Position) / range_
			c1 := parseHexColor(stops[i].Color)
			c2 := parseHexColor(stops[i+1].Color)
			return lerpColor(c1, c2, localT)
		}
	}

	return parseHexColor(stops[0].Color)
}

func lerpColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return color.NRGBA{
		R: uint8(float64(r1>>8)*(1-t) + float64(r2>>8)*t),
		G: uint8(float64(g1>>8)*(1-t) + float64(g2>>8)*t),
		B: uint8(float64(b1>>8)*(1-t) + float64(b2>>8)*t),
		A: uint8(float64(a1>>8)*(1-t) + float64(a2>>8)*t),
	}
}

func parseHexColor(hex string) color.Color {
	if hex == "" {
		return color.Transparent
	}
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b, a uint8
	a = 255

	switch len(hex) {
	case 3:
		fmt.Sscanf(hex, "%1x%1x%1x", &r, &g, &b)
		r *= 17
		g *= 17
		b *= 17
	case 4:
		fmt.Sscanf(hex, "%1x%1x%1x%1x", &r, &g, &b, &a)
		r *= 17
		g *= 17
		b *= 17
		a *= 17
	case 6:
		fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	case 8:
		fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
	default:
		return color.Black
	}

	return color.NRGBA{R: r, G: g, B: b, A: a}
}

func intOrDefault(val float64, def float64) int {
	if val > 0 {
		return int(val)
	}
	return int(def)
}

func axToAlign(ax float64) int {
	if ax >= 0.75 {
		return 2 // right
	}
	if ax >= 0.25 {
		return 1 // center
	}
	return 0 // left
}
