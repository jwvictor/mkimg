package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
