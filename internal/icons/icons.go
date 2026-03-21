package icons

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// IconDir returns the directory where icon fonts are cached.
func IconDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mkimg", "icons")
}

// EnsureIconDir creates the icon directory if needed.
func EnsureIconDir() error {
	return os.MkdirAll(IconDir(), 0755)
}

const (
	CollectionMaterial    = "material"
	CollectionFontAwesome = "fontawesome"
)

// MaterialIconMeta represents a Material Design icon entry.
type MaterialIconMeta struct {
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
	Codepoint  string   `json:"codepoint"`
}

// InstallMaterialIcons downloads the Material Symbols font.
func InstallMaterialIcons() error {
	if err := EnsureIconDir(); err != nil {
		return err
	}

	dir := filepath.Join(IconDir(), "material")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Download Material Symbols Outlined (variable font, includes all icons)
	fontURL := "https://github.com/google/material-design-icons/raw/master/variablefont/MaterialSymbolsOutlined%5BFILL%2CGRAD%2Copsz%2Cwght%5D.ttf"
	fmt.Println("  Downloading Material Symbols font...")
	resp, err := http.Get(fontURL)
	if err != nil {
		return fmt.Errorf("download Material Symbols: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read font data: %w", err)
	}

	fontPath := filepath.Join(dir, "MaterialSymbolsOutlined.ttf")
	if err := os.WriteFile(fontPath, data, 0644); err != nil {
		return fmt.Errorf("write font file: %w", err)
	}

	// Also download the codepoints mapping
	codepointsURL := "https://raw.githubusercontent.com/google/material-design-icons/master/variablefont/MaterialSymbolsOutlined%5BFILL%2CGRAD%2Copsz%2Cwght%5D.codepoints"
	fmt.Println("  Downloading codepoints mapping...")
	resp2, err := http.Get(codepointsURL)
	if err != nil {
		fmt.Printf("  Warning: could not download codepoints: %v\n", err)
		return nil // Non-fatal
	}
	defer resp2.Body.Close()

	if resp2.StatusCode == 200 {
		cpData, _ := io.ReadAll(resp2.Body)
		cpPath := filepath.Join(dir, "codepoints.txt")
		os.WriteFile(cpPath, cpData, 0644)
	}

	fmt.Println("  Material Symbols installed successfully")
	return nil
}

// InstallFontAwesome downloads Font Awesome Free.
func InstallFontAwesome() error {
	if err := EnsureIconDir(); err != nil {
		return err
	}

	dir := filepath.Join(IconDir(), "fontawesome")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Download Font Awesome Free Solid from CDN
	fonts := map[string]string{
		"fa-solid-900.ttf":   "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/webfonts/fa-solid-900.ttf",
		"fa-regular-400.ttf": "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/webfonts/fa-regular-400.ttf",
		"fa-brands-400.ttf":  "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/webfonts/fa-brands-400.ttf",
	}

	for name, url := range fonts {
		fmt.Printf("  Downloading %s...\n", name)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("download %s: %w", name, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("  Warning: could not download %s (status %d)\n", name, resp.StatusCode)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}

	// Download the metadata/cheatsheet for lookups
	metaURL := "https://raw.githubusercontent.com/FortAwesome/Font-Awesome/6.x/metadata/icons.json"
	fmt.Println("  Downloading icon metadata...")
	resp, err := http.Get(metaURL)
	if err != nil {
		fmt.Printf("  Warning: could not download metadata: %v\n", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			data, _ := io.ReadAll(resp.Body)
			os.WriteFile(filepath.Join(dir, "icons.json"), data, 0644)
		}
	}

	fmt.Println("  Font Awesome Free installed successfully")
	return nil
}

// LookupMaterialIcon returns the Unicode codepoint for a Material icon name.
func LookupMaterialIcon(name string) (rune, error) {
	cpPath := filepath.Join(IconDir(), "material", "codepoints.txt")
	data, err := os.ReadFile(cpPath)
	if err != nil {
		return 0, fmt.Errorf("codepoints not found — run: mkimg icon install material")
	}

	name = strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[0] == name {
			var cp rune
			fmt.Sscanf(parts[1], "%x", &cp)
			return cp, nil
		}
	}
	return 0, fmt.Errorf("icon %q not found in Material Symbols", name)
}

// LookupFontAwesomeIcon returns the Unicode codepoint for a Font Awesome icon.
func LookupFontAwesomeIcon(name string) (rune, string, error) {
	metaPath := filepath.Join(IconDir(), "fontawesome", "icons.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return 0, "", fmt.Errorf("metadata not found — run: mkimg icon install fontawesome")
	}

	var icons map[string]struct {
		Unicode string `json:"unicode"`
		Styles  []string `json:"styles"`
	}
	if err := json.Unmarshal(data, &icons); err != nil {
		return 0, "", fmt.Errorf("parse metadata: %w", err)
	}

	name = strings.ToLower(strings.TrimPrefix(name, "fa-"))
	if icon, ok := icons[name]; ok {
		var cp rune
		fmt.Sscanf(icon.Unicode, "%x", &cp)
		style := "solid"
		if len(icon.Styles) > 0 {
			style = icon.Styles[0]
		}
		return cp, style, nil
	}
	return 0, "", fmt.Errorf("icon %q not found in Font Awesome", name)
}

// SearchMaterialIcons searches for material icons by keyword.
func SearchMaterialIcons(query string) ([]string, error) {
	cpPath := filepath.Join(IconDir(), "material", "codepoints.txt")
	data, err := os.ReadFile(cpPath)
	if err != nil {
		return nil, fmt.Errorf("codepoints not found — run: mkimg icon install material")
	}

	query = strings.ToLower(query)
	var results []string
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 1 && strings.Contains(parts[0], query) {
			results = append(results, parts[0])
		}
	}
	return results, nil
}

// MaterialFontPath returns the path to the Material Symbols font file.
func MaterialFontPath() string {
	return filepath.Join(IconDir(), "material", "MaterialSymbolsOutlined.ttf")
}

// FontAwesomeFontPath returns the path to a Font Awesome font file.
func FontAwesomeFontPath(style string) string {
	switch style {
	case "regular":
		return filepath.Join(IconDir(), "fontawesome", "fa-regular-400.ttf")
	case "brands":
		return filepath.Join(IconDir(), "fontawesome", "fa-brands-400.ttf")
	default:
		return filepath.Join(IconDir(), "fontawesome", "fa-solid-900.ttf")
	}
}

// ListCollections returns available icon collections.
func ListCollections() []string {
	var collections []string
	dir := IconDir()
	if entries, err := os.ReadDir(dir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				collections = append(collections, e.Name())
			}
		}
	}
	return collections
}
