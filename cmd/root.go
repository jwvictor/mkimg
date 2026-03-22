package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags.
var version = "dev"

// updateWarning is populated by the background update check goroutine.
var updateWarning string
var updateDone = make(chan struct{}, 1)
var updateOnce sync.Once

var rootCmd = &cobra.Command{
	Use:   "mkimg",
	Short: "CLI image editor and generator",
	Long: `mkimg — a command-line image editor for creating ad creatives,
social media graphics, and more. Supports layering, AI generation,
Google Fonts, icon libraries, filters, and effects.

Workflow:
  1. mkimg new my-ad --preset instagram-story
  2. mkimg layer add gradient --from "#667eea" --to "#764ba2" --angle 135
  3. mkimg layer add text --content "Summer Sale" --font Montserrat --size 72
  4. mkimg layer add ai --prompt "minimalist product photo"
  5. mkimg filter <layer-id> blur --radius 3
  6. mkimg render -o my-ad.png`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if version == "dev" {
			return
		}
		// Skip the check for version/update commands — they handle it themselves
		name := cmd.Name()
		if name == "version" || name == "update" {
			return
		}
		go backgroundUpdateCheck()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Wait for the background check to finish, but don't block for more than 2s
		select {
		case <-updateDone:
		case <-time.After(2 * time.Second):
		}
		if updateWarning != "" {
			fmt.Fprintln(os.Stderr, updateWarning)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringP("project", "p", "", "Project file (auto-detected if omitted)")
}

// backgroundUpdateCheck checks for a newer release, throttled to once per 24h.
// Results are cached in ~/.mkimg/last_update_check.
func backgroundUpdateCheck() {
	defer func() { updateDone <- struct{}{} }()
	updateOnce.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		cacheDir := filepath.Join(home, ".mkimg")
		cacheFile := filepath.Join(cacheDir, "last_update_check")

		// Check if we already checked recently
		if data, err := os.ReadFile(cacheFile); err == nil {
			parts := strings.SplitN(string(data), "\n", 2)
			if len(parts) == 2 {
				if ts, err := time.Parse(time.RFC3339, parts[0]); err == nil {
					if time.Since(ts) < 24*time.Hour {
						// Use cached result
						latest := strings.TrimSpace(parts[1])
						if latest != "" && latest != version {
							updateWarning = fmt.Sprintf("\nmkimg %s is available (you have %s). Run 'mkimg update' to upgrade.", latest, version)
						}
						return
					}
				}
			}
		}

		// Fetch from GitHub
		rel, err := fetchLatestRelease()
		if err != nil {
			return // fail silently
		}

		latest := strings.TrimPrefix(rel.TagName, "v")

		// Cache the result
		os.MkdirAll(cacheDir, 0755)
		cacheData := time.Now().Format(time.RFC3339) + "\n" + latest
		os.WriteFile(cacheFile, []byte(cacheData), 0644)

		if latest != version {
			updateWarning = fmt.Sprintf("\nmkimg %s is available (you have %s). Run 'mkimg update' to upgrade.", latest, version)
		}
	})
}
