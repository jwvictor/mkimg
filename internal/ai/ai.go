package ai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

// GenerateImage uses Gemini (Nano Banana) to generate an image from a prompt.
func GenerateImage(prompt string, outputPath string, opts Options) error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY or GOOGLE_API_KEY environment variable required")
	}

	model := opts.Model
	if model == "" {
		model = "gemini-3.1-flash-image-preview"
	}

	// Build the request
	contents := []ContentPart{}

	// Add reference image if provided
	if opts.ReferenceImage != "" {
		imgData, mimeType, err := loadImageAsBase64(opts.ReferenceImage)
		if err != nil {
			return fmt.Errorf("load reference image: %w", err)
		}
		contents = append(contents, ContentPart{
			InlineData: &InlineData{
				MimeType: mimeType,
				Data:     imgData,
			},
		})
	}

	// Add the text prompt
	contents = append(contents, ContentPart{
		Text: prompt,
	})

	reqBody := GeminiRequest{
		Contents: []Content{
			{Parts: contents},
		},
		GenerationConfig: GenerationConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
		},
	}

	if opts.AspectRatio != "" {
		reqBody.GenerationConfig.ImageConfig = &ImageConfig{
			AspectRatio: opts.AspectRatio,
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	// Extract the image from the response
	for _, candidate := range geminiResp.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && strings.HasPrefix(part.InlineData.MimeType, "image/") {
				imgData, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
				if err != nil {
					return fmt.Errorf("decode image data: %w", err)
				}
				if err := os.WriteFile(outputPath, imgData, 0644); err != nil {
					return fmt.Errorf("write output: %w", err)
				}
				fmt.Printf("  Generated image saved to %s\n", outputPath)
				return nil
			}
		}
	}

	// Check if there's text content to report
	for _, candidate := range geminiResp.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				return fmt.Errorf("model returned text instead of image: %s", part.Text)
			}
		}
	}

	return fmt.Errorf("no image found in API response")
}

// GenerateToImage generates an AI image and returns it as an image.Image.
func GenerateToImage(prompt string, opts Options) (image.Image, error) {
	// Generate to a temp file, then load
	tmpFile, err := os.CreateTemp("", "mkimg-ai-*.png")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := GenerateImage(prompt, tmpPath, opts); err != nil {
		return nil, err
	}

	f, err := os.Open(tmpPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func loadImageAsBase64(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	// Detect MIME type
	mimeType := "image/png"
	if strings.HasSuffix(strings.ToLower(path), ".jpg") || strings.HasSuffix(strings.ToLower(path), ".jpeg") {
		mimeType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(path), ".webp") {
		mimeType = "image/webp"
	}

	return base64.StdEncoding.EncodeToString(data), mimeType, nil
}

// SaveImage saves an image.Image as PNG.
func SaveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// Options for AI image generation.
type Options struct {
	Model          string
	ReferenceImage string
	AspectRatio    string
}

// Gemini API types

type GeminiRequest struct {
	Contents         []Content        `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

type Content struct {
	Parts []ContentPart `json:"parts"`
}

type ContentPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GenerationConfig struct {
	ResponseModalities []string     `json:"responseModalities"`
	ImageConfig        *ImageConfig `json:"imageConfig,omitempty"`
}

type ImageConfig struct {
	AspectRatio string `json:"aspectRatio,omitempty"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}
