package fonts

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// FontDir returns the directory where fonts are cached.
func FontDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mkimg", "fonts")
}

// EnsureFontDir creates the font directory if it doesn't exist.
func EnsureFontDir() error {
	return os.MkdirAll(FontDir(), 0755)
}

// InstallFont downloads a font family directly from Google Fonts (no API key needed).
// It uses the CSS2 API to discover font file URLs and downloads them directly.
func InstallFont(family string) error {
	if err := EnsureFontDir(); err != nil {
		return err
	}

	familyDir := filepath.Join(FontDir(), sanitizeName(family))
	if err := os.MkdirAll(familyDir, 0755); err != nil {
		return err
	}

	// Use the CSS2 API to get font URLs (no API key needed).
	// We request all standard weights. The API ignores weights the font doesn't have.
	// Fetch normal weights first, then italic separately.
	weights := "100;200;300;400;500;600;700;800;900"
	cssURL := fmt.Sprintf(
		"https://fonts.googleapis.com/css2?family=%s:wght@%s&display=swap",
		url.QueryEscape(family), weights,
	)

	cssData, err := fetchFontCSS(cssURL)
	if err != nil {
		return fmt.Errorf("font %q not found on Google Fonts: %w", family, err)
	}

	// Also try to fetch italic variants
	italicURL := fmt.Sprintf(
		"https://fonts.googleapis.com/css2?family=%s:ital,wght@1,100;1,200;1,300;1,400;1,500;1,600;1,700;1,800;1,900&display=swap",
		url.QueryEscape(family),
	)
	if italicCSS, err := fetchFontCSS(italicURL); err == nil {
		cssData = append(cssData, italicCSS...)
	}

	// Parse out font URLs and metadata from the CSS
	type fontEntry struct {
		url    string
		weight string
		style  string
		subset string
	}

	var entries []fontEntry
	blocks := strings.Split(string(cssData), "@font-face")

	urlRe := regexp.MustCompile(`url\(([^)]+)\)`)
	weightRe := regexp.MustCompile(`font-weight:\s*(\d+)`)
	styleRe := regexp.MustCompile(`font-style:\s*(\w+)`)
	subsetRe := regexp.MustCompile(`/\*\s*([\w-]+)\s*\*/`)

	for _, block := range blocks {
		urlMatch := urlRe.FindStringSubmatch(block)
		if urlMatch == nil {
			continue
		}
		fontURL := urlMatch[1]

		weight := "400"
		if m := weightRe.FindStringSubmatch(block); m != nil {
			weight = m[1]
		}
		style := "normal"
		if m := styleRe.FindStringSubmatch(block); m != nil {
			style = m[1]
		}
		subset := "latin"
		if m := subsetRe.FindStringSubmatch(block); m != nil {
			subset = m[1]
		}

		// Only download latin subset to keep things small
		if subset != "latin" {
			continue
		}

		entries = append(entries, fontEntry{
			url:    fontURL,
			weight: weight,
			style:  style,
			subset: subset,
		})
	}

	if len(entries) == 0 {
		return fmt.Errorf("no font files found for %q", family)
	}

	// Deduplicate by weight+style (keep first match, which is latin)
	seen := map[string]bool{}
	installed := 0

	weightNames := map[string]string{
		"100": "Thin", "200": "ExtraLight", "300": "Light",
		"400": "Regular", "500": "Medium", "600": "SemiBold",
		"700": "Bold", "800": "ExtraBold", "900": "Black",
	}

	for _, e := range entries {
		key := e.weight + "-" + e.style
		if seen[key] {
			continue
		}
		seen[key] = true

		wName := weightNames[e.weight]
		if wName == "" {
			wName = e.weight
		}
		suffix := wName
		if e.style == "italic" {
			suffix += "Italic"
		}

		filename := fmt.Sprintf("%s-%s.ttf", sanitizeName(family), suffix)
		fmt.Printf("  %s %s", family, suffix)

		fontResp, err := http.Get(e.url)
		if err != nil {
			fmt.Printf(" (failed: %v)\n", err)
			continue
		}
		data, err := io.ReadAll(fontResp.Body)
		fontResp.Body.Close()
		if err != nil {
			fmt.Printf(" (failed: %v)\n", err)
			continue
		}

		outPath := filepath.Join(familyDir, filename)
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", filename, err)
		}
		fmt.Printf("\n")
		installed++
	}

	if installed == 0 {
		return fmt.Errorf("failed to download any files for %q", family)
	}

	fmt.Printf("  Installed %s (%d variants) → %s\n", family, installed, familyDir)
	return nil
}

// SearchGoogleFonts searches for fonts. Uses the API key if available,
// otherwise falls back to the CSS API (no key needed).
func fetchFontCSS(cssURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", cssURL, nil)
	if err != nil {
		return nil, err
	}
	// A simple/unknown user-agent makes Google serve raw TTF (truetype) format,
	// which golang.org/x/image/font/opentype can parse. Modern browser UAs get woff/woff2.
	req.Header.Set("User-Agent", "mkimg/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func SearchGoogleFonts(query string) ([]FontResult, error) {
	// Try the metadata endpoint first (no key needed)
	results, err := searchViaCSS(query)
	if err == nil && len(results) > 0 {
		return results, nil
	}

	// Fallback: try the webfonts API if an API key is available
	apiKey := os.Getenv("GOOGLE_FONTS_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey != "" {
		return searchViaAPI(query, apiKey)
	}

	if err != nil {
		return nil, err
	}
	return results, nil
}

// FontResult is a simplified font search result.
type FontResult struct {
	Family   string
	Category string
	Variants int
}

// searchViaCSS uses the Google Fonts CSS2 API to check if fonts exist.
// For search/listing, we use the metadata endpoint.
func searchViaCSS(query string) ([]FontResult, error) {
	// The Google Fonts metadata API doesn't require a key
	metaURL := "https://fonts.google.com/metadata/fonts"
	resp, err := http.Get(metaURL)
	if err != nil {
		return nil, fmt.Errorf("fetch font metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("metadata API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read metadata: %w", err)
	}

	// The response starts with ")]}'" which is an XSS prevention prefix — strip it
	bodyStr := string(body)
	if idx := strings.Index(bodyStr, "\n"); idx >= 0 {
		bodyStr = bodyStr[idx+1:]
	}

	var meta struct {
		FamilyMetadataList []struct {
			Family   string            `json:"family"`
			Category string            `json:"category"`
			Fonts    map[string]interface{} `json:"fonts"`
		} `json:"familyMetadataList"`
	}
	if err := json.Unmarshal([]byte(bodyStr), &meta); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}

	query = strings.ToLower(query)
	var results []FontResult
	for _, f := range meta.FamilyMetadataList {
		if query == "" || strings.Contains(strings.ToLower(f.Family), query) {
			results = append(results, FontResult{
				Family:   f.Family,
				Category: f.Category,
				Variants: len(f.Fonts),
			})
		}
	}
	return results, nil
}

func searchViaAPI(query, apiKey string) ([]FontResult, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/webfonts/v1/webfonts?key=%s&sort=popularity", apiKey)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("fetch Google Fonts API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			Family   string            `json:"family"`
			Category string            `json:"category"`
			Variants []string          `json:"variants"`
			Files    map[string]string `json:"files"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse API response: %w", err)
	}

	query = strings.ToLower(query)
	var results []FontResult
	for _, f := range result.Items {
		if query == "" || strings.Contains(strings.ToLower(f.Family), query) {
			results = append(results, FontResult{
				Family:   f.Family,
				Category: f.Category,
				Variants: len(f.Variants),
			})
		}
	}
	return results, nil
}

// ListInstalled returns all installed font families.
func ListInstalled() ([]string, error) {
	dir := FontDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var families []string
	for _, e := range entries {
		if e.IsDir() {
			families = append(families, e.Name())
		}
	}
	sort.Strings(families)
	return families, nil
}

// LoadFace loads a font face for rendering.
func LoadFace(family string, size float64, variant string) (font.Face, error) {
	if variant == "" {
		variant = "regular"
	}

	fontPath := findFontFile(family, variant)
	if fontPath == "" {
		return nil, fmt.Errorf("font %q (variant %q) not found — install with: mkimg font install %q", family, variant, family)
	}

	data, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("read font file: %w", err)
	}

	f, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("create font face: %w", err)
	}

	return face, nil
}

// findFontFile locates a font file for the given family and variant.
func findFontFile(family, variant string) string {
	dir := filepath.Join(FontDir(), sanitizeName(family))

	// Try exact match patterns common in Google Fonts downloads
	variantMap := map[string][]string{
		"regular":    {"Regular", "regular"},
		"bold":       {"Bold", "bold"},
		"700":        {"Bold", "bold", "700"},
		"italic":     {"Italic", "italic"},
		"bolditalic": {"BoldItalic", "bolditalic"},
		"100":        {"Thin", "thin", "100"},
		"200":        {"ExtraLight", "extralight", "200"},
		"300":        {"Light", "light", "300"},
		"500":        {"Medium", "medium", "500"},
		"600":        {"SemiBold", "semibold", "600"},
		"800":        {"ExtraBold", "extrabold", "800"},
		"900":        {"Black", "black", "900"},
	}

	suffixes := variantMap[variant]
	if suffixes == nil {
		suffixes = []string{variant}
	}

	sanitized := sanitizeName(family)
	noSpaces := strings.ReplaceAll(family, " ", "")

	// Check mkimg font cache with various naming patterns
	for _, suffix := range suffixes {
		candidates := []string{
			filepath.Join(dir, fmt.Sprintf("%s-%s.ttf", sanitized, suffix)),
			filepath.Join(dir, fmt.Sprintf("%s-%s.otf", sanitized, suffix)),
			filepath.Join(dir, fmt.Sprintf("%s-%s.woff", sanitized, suffix)),
			filepath.Join(dir, fmt.Sprintf("%s-%s.ttf", noSpaces, suffix)),
			filepath.Join(dir, fmt.Sprintf("%s-%s.otf", noSpaces, suffix)),
			filepath.Join(dir, fmt.Sprintf("%s-%s.woff", noSpaces, suffix)),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	}

	// If variant is regular, also try just the family name
	if variant == "regular" {
		for _, ext := range []string{".ttf", ".otf", ".woff"} {
			c := filepath.Join(dir, sanitized+ext)
			if _, err := os.Stat(c); err == nil {
				return c
			}
			c = filepath.Join(dir, noSpaces+ext)
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	}

	// Last resort: glob for anything with the variant name in the font dir
	if entries, err := os.ReadDir(dir); err == nil {
		varLower := strings.ToLower(variant)
		// Map common variant names
		if varLower == "700" || varLower == "bold" {
			varLower = "bold"
		}
		if varLower == "regular" {
			varLower = "regular"
		}

		re := regexp.MustCompile(`(?i)[-_]` + regexp.QuoteMeta(varLower) + `\.(ttf|otf|woff)$`)
		for _, e := range entries {
			if re.MatchString(e.Name()) {
				return filepath.Join(dir, e.Name())
			}
		}

		// If looking for regular variant, try the first .ttf file that doesn't have
		// a weight suffix (Bold, Italic, etc.)
		if variant == "regular" {
			reWeight := regexp.MustCompile(`(?i)(bold|italic|thin|light|medium|semi|extra|black)`)
			for _, e := range entries {
				name := e.Name()
				ext := strings.ToLower(filepath.Ext(name))
				if (ext == ".ttf" || ext == ".otf" || ext == ".woff") && !reWeight.MatchString(name) {
					return filepath.Join(dir, name)
				}
			}
			// Still nothing? Just return the first font file
			for _, e := range entries {
				ext := strings.ToLower(filepath.Ext(e.Name()))
				if ext == ".ttf" || ext == ".otf" || ext == ".woff" {
					return filepath.Join(dir, e.Name())
				}
			}
		}
	}

	// Check system fonts
	for _, sysDir := range systemFontDirs() {
		for _, suffix := range suffixes {
			candidates := []string{
				filepath.Join(sysDir, family+".ttf"),
				filepath.Join(sysDir, family+"-"+suffix+".ttf"),
				filepath.Join(sysDir, noSpaces+"-"+suffix+".ttf"),
			}
			for _, c := range candidates {
				if _, err := os.Stat(c); err == nil {
					return c
				}
			}
		}
	}

	return ""
}

func systemFontDirs() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/System/Library/Fonts",
			"/Library/Fonts",
			filepath.Join(os.Getenv("HOME"), "Library/Fonts"),
		}
	case "linux":
		return []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			filepath.Join(os.Getenv("HOME"), ".fonts"),
			filepath.Join(os.Getenv("HOME"), ".local/share/fonts"),
		}
	default:
		return nil
	}
}

func sanitizeName(name string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		if r == ' ' {
			return '-'
		}
		return -1
	}, name)
}
