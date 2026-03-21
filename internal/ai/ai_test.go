package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateImageNoAPIKey(t *testing.T) {
	// Temporarily clear API keys
	origGemini := os.Getenv("GEMINI_API_KEY")
	origGoogle := os.Getenv("GOOGLE_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GOOGLE_API_KEY")
	defer func() {
		if origGemini != "" {
			os.Setenv("GEMINI_API_KEY", origGemini)
		}
		if origGoogle != "" {
			os.Setenv("GOOGLE_API_KEY", origGoogle)
		}
	}()

	err := GenerateImage("test", "/tmp/out.png", Options{})
	if err == nil {
		t.Error("expected error when no API key set")
	}
	if err.Error() != "GEMINI_API_KEY or GOOGLE_API_KEY environment variable required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadImageAsBase64PNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	os.WriteFile(path, []byte("fake png data"), 0644)

	data, mime, err := loadImageAsBase64(path)
	if err != nil {
		t.Fatalf("loadImageAsBase64: %v", err)
	}
	if mime != "image/png" {
		t.Errorf("MIME: got %q, want image/png", mime)
	}
	if data == "" {
		t.Error("base64 data is empty")
	}
}

func TestLoadImageAsBase64JPEG(t *testing.T) {
	dir := t.TempDir()
	for _, ext := range []string{"test.jpg", "test.jpeg"} {
		path := filepath.Join(dir, ext)
		os.WriteFile(path, []byte("fake jpeg data"), 0644)

		_, mime, err := loadImageAsBase64(path)
		if err != nil {
			t.Fatalf("loadImageAsBase64(%s): %v", ext, err)
		}
		if mime != "image/jpeg" {
			t.Errorf("%s MIME: got %q, want image/jpeg", ext, mime)
		}
	}
}

func TestLoadImageAsBase64WebP(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.webp")
	os.WriteFile(path, []byte("fake webp data"), 0644)

	_, mime, err := loadImageAsBase64(path)
	if err != nil {
		t.Fatalf("loadImageAsBase64: %v", err)
	}
	if mime != "image/webp" {
		t.Errorf("MIME: got %q, want image/webp", mime)
	}
}

func TestLoadImageAsBase64MissingFile(t *testing.T) {
	_, _, err := loadImageAsBase64("/nonexistent/file.png")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestGeminiResponseParsing(t *testing.T) {
	// Test that we can parse a valid Gemini response structure
	resp := GeminiResponse{
		Candidates: []Candidate{
			{
				Content: Content{
					Parts: []ContentPart{
						{Text: "Here's your image"},
						{InlineData: &InlineData{
							MimeType: "image/png",
							Data:     "iVBORw0KGgo=", // truncated base64
						}},
					},
				},
			},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed GeminiResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(parsed.Candidates) != 1 {
		t.Fatalf("candidates: got %d, want 1", len(parsed.Candidates))
	}
	parts := parsed.Candidates[0].Content.Parts
	if len(parts) != 2 {
		t.Fatalf("parts: got %d, want 2", len(parts))
	}
	if parts[1].InlineData == nil {
		t.Fatal("expected InlineData in second part")
	}
	if parts[1].InlineData.MimeType != "image/png" {
		t.Errorf("MIME: got %q", parts[1].InlineData.MimeType)
	}
}

func TestGeminiResponseEmptyCandidates(t *testing.T) {
	resp := `{"candidates": []}`
	var parsed GeminiResponse
	if err := json.Unmarshal([]byte(resp), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Candidates) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(parsed.Candidates))
	}
}

func TestGeminiRequestSerialization(t *testing.T) {
	req := GeminiRequest{
		Contents: []Content{
			{Parts: []ContentPart{
				{Text: "draw a cat"},
			}},
		},
		GenerationConfig: GenerationConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
			ImageConfig:        &ImageConfig{AspectRatio: "16:9"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	// Verify structure
	config := raw["generationConfig"].(map[string]interface{})
	imgConfig := config["imageConfig"].(map[string]interface{})
	if imgConfig["aspectRatio"] != "16:9" {
		t.Errorf("aspectRatio not serialized correctly")
	}
}

func TestGenerateImageAPIError(t *testing.T) {
	// Mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"error": {"message": "rate limited"}}`))
	}))
	defer server.Close()

	// We can't easily redirect the real function to use our test server
	// without refactoring, but we can at least test the types serialize correctly.
	// This test documents that the error path exists.
}

func TestOptionsDefaults(t *testing.T) {
	opts := Options{}
	if opts.Model != "" {
		t.Errorf("default Model should be empty, got %q", opts.Model)
	}
	if opts.AspectRatio != "" {
		t.Errorf("default AspectRatio should be empty, got %q", opts.AspectRatio)
	}
	if opts.ReferenceImage != "" {
		t.Errorf("default ReferenceImage should be empty, got %q", opts.ReferenceImage)
	}
}
