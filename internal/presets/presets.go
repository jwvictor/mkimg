package presets

import "github.com/jwvictor/mkimg/internal/project"

// Preset defines a canvas template with optional starter layers.
type Preset struct {
	Name        string
	Description string
	Width       int
	Height      int
	Background  string
	Layers      []project.Layer
}

// All returns all available presets.
func All() []Preset {
	return []Preset{
		// Social Media
		{
			Name:        "instagram-post",
			Description: "Instagram post (1080x1080)",
			Width:       1080,
			Height:      1080,
			Background:  "#ffffff",
		},
		{
			Name:        "instagram-story",
			Description: "Instagram/TikTok story (1080x1920)",
			Width:       1080,
			Height:      1920,
			Background:  "#000000",
		},
		{
			Name:        "instagram-reel",
			Description: "Instagram Reel cover (1080x1920)",
			Width:       1080,
			Height:      1920,
			Background:  "#1a1a2e",
		},
		{
			Name:        "facebook-post",
			Description: "Facebook post (1200x630)",
			Width:       1200,
			Height:      630,
			Background:  "#ffffff",
		},
		{
			Name:        "facebook-cover",
			Description: "Facebook cover photo (820x312)",
			Width:       820,
			Height:      312,
			Background:  "#f0f2f5",
		},
		{
			Name:        "twitter-post",
			Description: "Twitter/X post image (1600x900)",
			Width:       1600,
			Height:      900,
			Background:  "#ffffff",
		},
		{
			Name:        "twitter-header",
			Description: "Twitter/X header (1500x500)",
			Width:       1500,
			Height:      500,
			Background:  "#1da1f2",
		},
		{
			Name:        "linkedin-post",
			Description: "LinkedIn post (1200x627)",
			Width:       1200,
			Height:      627,
			Background:  "#ffffff",
		},
		{
			Name:        "youtube-thumbnail",
			Description: "YouTube thumbnail (1280x720)",
			Width:       1280,
			Height:      720,
			Background:  "#ff0000",
		},
		{
			Name:        "pinterest-pin",
			Description: "Pinterest pin (1000x1500)",
			Width:       1000,
			Height:      1500,
			Background:  "#ffffff",
		},

		// Ads
		{
			Name:        "google-display",
			Description: "Google Display ad (300x250)",
			Width:       300,
			Height:      250,
			Background:  "#ffffff",
		},
		{
			Name:        "leaderboard",
			Description: "Leaderboard ad (728x90)",
			Width:       728,
			Height:      90,
			Background:  "#ffffff",
		},
		{
			Name:        "wide-skyscraper",
			Description: "Wide skyscraper ad (160x600)",
			Width:       160,
			Height:      600,
			Background:  "#ffffff",
		},

		// Print / General
		{
			Name:        "business-card",
			Description: "Business card (1050x600 @ 300dpi)",
			Width:       1050,
			Height:      600,
			Background:  "#ffffff",
		},
		{
			Name:        "poster-24x36",
			Description: "Poster 24x36 inches (7200x10800 @ 300dpi)",
			Width:       7200,
			Height:      10800,
			Background:  "#ffffff",
		},
		{
			Name:        "a4",
			Description: "A4 page (2480x3508 @ 300dpi)",
			Width:       2480,
			Height:      3508,
			Background:  "#ffffff",
		},
		{
			Name:        "presentation",
			Description: "Presentation slide (1920x1080)",
			Width:       1920,
			Height:      1080,
			Background:  "#1a1a2e",
		},
		{
			Name:        "og-image",
			Description: "Open Graph / social share image (1200x630)",
			Width:       1200,
			Height:      630,
			Background:  "#ffffff",
		},
		{
			Name:        "app-icon",
			Description: "App icon (1024x1024)",
			Width:       1024,
			Height:      1024,
			Background:  "#ffffff",
		},
		{
			Name:        "favicon",
			Description: "Favicon (512x512)",
			Width:       512,
			Height:      512,
			Background:  "#ffffff",
		},
	}
}

// Get returns a preset by name.
func Get(name string) *Preset {
	for _, p := range All() {
		if p.Name == name {
			return &p
		}
	}
	return nil
}
