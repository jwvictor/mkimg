package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
)

// Apply applies a named filter with params to an image.
func Apply(img image.Image, filterType string, params map[string]float64) (image.Image, error) {
	switch filterType {
	case "blur":
		sigma := getParam(params, "radius", 3.0)
		return imaging.Blur(img, sigma), nil

	case "sharpen":
		sigma := getParam(params, "radius", 1.0)
		return imaging.Sharpen(img, sigma), nil

	case "brightness":
		v := getParam(params, "value", 0)
		return imaging.AdjustBrightness(img, v), nil

	case "contrast":
		v := getParam(params, "value", 0)
		return imaging.AdjustContrast(img, v), nil

	case "saturation":
		v := getParam(params, "value", 0)
		return imaging.AdjustSaturation(img, v), nil

	case "gamma":
		v := getParam(params, "value", 1.0)
		return imaging.AdjustGamma(img, v), nil

	case "hue":
		v := getParam(params, "value", 0)
		return adjustHue(img, v), nil

	case "grayscale":
		return imaging.Grayscale(img), nil

	case "sepia":
		return applySepia(img), nil

	case "invert":
		return imaging.Invert(img), nil

	case "pixelate":
		size := int(getParam(params, "size", 10))
		if size < 1 {
			size = 1
		}
		bounds := img.Bounds()
		w, h := bounds.Dx(), bounds.Dy()
		small := imaging.Resize(img, w/size, h/size, imaging.NearestNeighbor)
		return imaging.Resize(small, w, h, imaging.NearestNeighbor), nil

	case "vignette":
		strength := getParam(params, "strength", 0.5)
		return applyVignette(img, strength), nil

	case "noise":
		amount := getParam(params, "amount", 20)
		return applyNoise(img, amount), nil

	case "posterize":
		levels := int(getParam(params, "levels", 4))
		return applyPosterize(img, levels), nil

	case "emboss":
		return applyConvolution(img, [9]float64{-2, -1, 0, -1, 1, 1, 0, 1, 2}), nil

	case "edge":
		return applyConvolution(img, [9]float64{-1, -1, -1, -1, 8, -1, -1, -1, -1}), nil

	case "glow":
		radius := getParam(params, "radius", 5.0)
		strength := getParam(params, "strength", 0.5)
		return applyGlow(img, radius, strength), nil

	case "duotone":
		// Expects "color1" and "color2" as hex encoded in params as RGB ints
		return applyDuotone(img,
			getParam(params, "r1", 0), getParam(params, "g1", 0), getParam(params, "b1", 50),
			getParam(params, "r2", 255), getParam(params, "g2", 200), getParam(params, "b2", 100),
		), nil

	default:
		return nil, fmt.Errorf("unknown filter: %s", filterType)
	}
}

// ListFilters returns all available filter names with descriptions.
func ListFilters() []FilterInfo {
	return []FilterInfo{
		{"blur", "Gaussian blur", "radius (default: 3)"},
		{"sharpen", "Sharpen image", "radius (default: 1)"},
		{"brightness", "Adjust brightness", "value: -100 to 100 (default: 0)"},
		{"contrast", "Adjust contrast", "value: -100 to 100 (default: 0)"},
		{"saturation", "Adjust saturation", "value: -100 to 100 (default: 0)"},
		{"gamma", "Adjust gamma", "value (default: 1.0)"},
		{"hue", "Rotate hue", "value: -180 to 180 degrees"},
		{"grayscale", "Convert to grayscale", "no params"},
		{"sepia", "Apply sepia tone", "no params"},
		{"invert", "Invert colors", "no params"},
		{"pixelate", "Pixelate effect", "size (default: 10)"},
		{"vignette", "Vignette effect", "strength: 0 to 1 (default: 0.5)"},
		{"noise", "Add noise", "amount (default: 20)"},
		{"posterize", "Reduce color levels", "levels (default: 4)"},
		{"emboss", "Emboss effect", "no params"},
		{"edge", "Edge detection", "no params"},
		{"glow", "Soft glow", "radius (default: 5), strength: 0-1 (default: 0.5)"},
		{"duotone", "Two-tone effect", "r1,g1,b1,r2,g2,b2"},
	}
}

type FilterInfo struct {
	Name        string
	Description string
	Params      string
}

func getParam(params map[string]float64, key string, def float64) float64 {
	if v, ok := params[key]; ok {
		return v
	}
	return def
}

func applySepia(img image.Image) image.Image {
	bounds := img.Bounds()
	dst := imaging.Clone(img)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rf, gf, bf := float64(r)/65535.0, float64(g)/65535.0, float64(b)/65535.0
			nr := math.Min(1.0, 0.393*rf+0.769*gf+0.189*bf)
			ng := math.Min(1.0, 0.349*rf+0.686*gf+0.168*bf)
			nb := math.Min(1.0, 0.272*rf+0.534*gf+0.131*bf)
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(nr * 255),
				G: uint8(ng * 255),
				B: uint8(nb * 255),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func applyVignette(img image.Image, strength float64) image.Image {
	bounds := img.Bounds()
	dst := imaging.Clone(img)
	cx, cy := float64(bounds.Dx())/2.0, float64(bounds.Dy())/2.0
	maxDist := math.Sqrt(cx*cx + cy*cy)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dx := float64(x-bounds.Min.X) - cx
			dy := float64(y-bounds.Min.Y) - cy
			dist := math.Sqrt(dx*dx+dy*dy) / maxDist
			factor := 1.0 - strength*dist*dist
			if factor < 0 {
				factor = 0
			}

			r, g, b, a := img.At(x, y).RGBA()
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(float64(r>>8) * factor),
				G: uint8(float64(g>>8) * factor),
				B: uint8(float64(b>>8) * factor),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func applyNoise(img image.Image, amount float64) image.Image {
	bounds := img.Bounds()
	dst := imaging.Clone(img)
	// Simple deterministic-ish noise using position-based hash
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// Simple hash for noise
			hash := float64((x*2654435761+y*2246822519)%256) / 255.0
			noise := (hash - 0.5) * amount * 2
			nr := clampUint8(float64(r>>8) + noise)
			ng := clampUint8(float64(g>>8) + noise)
			nb := clampUint8(float64(b>>8) + noise)
			dst.SetNRGBA(x, y, color.NRGBA{R: nr, G: ng, B: nb, A: uint8(a >> 8)})
		}
	}
	return dst
}

func applyPosterize(img image.Image, levels int) image.Image {
	if levels < 2 {
		levels = 2
	}
	bounds := img.Bounds()
	dst := imaging.Clone(img)
	step := 255.0 / float64(levels-1)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(math.Round(float64(r>>8)/step) * step),
				G: uint8(math.Round(float64(g>>8)/step) * step),
				B: uint8(math.Round(float64(b>>8)/step) * step),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func applyConvolution(img image.Image, kernel [9]float64) image.Image {
	bounds := img.Bounds()
	dst := imaging.Clone(img)
	w, h := bounds.Dx(), bounds.Dy()

	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			var rs, gs, bs float64
			idx := 0
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					r, g, b, _ := img.At(bounds.Min.X+x+kx, bounds.Min.Y+y+ky).RGBA()
					rs += float64(r>>8) * kernel[idx]
					gs += float64(g>>8) * kernel[idx]
					bs += float64(b>>8) * kernel[idx]
					idx++
				}
			}
			_, _, _, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			dst.SetNRGBA(bounds.Min.X+x, bounds.Min.Y+y, color.NRGBA{
				R: clampUint8(rs),
				G: clampUint8(gs),
				B: clampUint8(bs),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func applyGlow(img image.Image, radius, strength float64) image.Image {
	blurred := imaging.Blur(img, radius)
	return imaging.OverlayCenter(img, imaging.AdjustBrightness(blurred, strength*50), 0.5)
}

func applyDuotone(img image.Image, r1, g1, b1, r2, g2, b2 float64) image.Image {
	gray := imaging.Grayscale(img)
	bounds := gray.Bounds()
	dst := imaging.Clone(gray)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, a := gray.At(x, y).RGBA()
			t := float64(r>>8) / 255.0
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r1*(1-t) + r2*t),
				G: uint8(g1*(1-t) + g2*t),
				B: uint8(b1*(1-t) + b2*t),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func adjustHue(img image.Image, degrees float64) image.Image {
	bounds := img.Bounds()
	dst := imaging.Clone(img)
	shift := degrees / 360.0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rf, gf, bf := float64(r)/65535.0, float64(g)/65535.0, float64(b)/65535.0
			h, s, l := rgbToHSL(rf, gf, bf)
			h += shift
			if h > 1 {
				h -= 1
			}
			if h < 0 {
				h += 1
			}
			nr, ng, nb := hslToRGB(h, s, l)
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(nr * 255),
				G: uint8(ng * 255),
				B: uint8(nb * 255),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2

	if max == min {
		return 0, 0, l
	}

	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h /= 6
	return
}

func hslToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		return l, l, l
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	r = hueToRGB(p, q, h+1.0/3.0)
	g = hueToRGB(p, q, h)
	b = hueToRGB(p, q, h-1.0/3.0)
	return
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func clampUint8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
