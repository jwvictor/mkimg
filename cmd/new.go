package cmd

import (
	"fmt"

	"github.com/jwvictor/mkimg/internal/presets"
	"github.com/jwvictor/mkimg/internal/project"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new project",
	Long: `Create a new mkimg project file.

Examples:
  mkimg new my-ad --width 1080 --height 1920 --bg "#1a1a2e"
  mkimg new promo --preset instagram-story
  mkimg new banner --preset youtube-thumbnail`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		presetName, _ := cmd.Flags().GetString("preset")
		width, _ := cmd.Flags().GetInt("width")
		height, _ := cmd.Flags().GetInt("height")
		bg, _ := cmd.Flags().GetString("bg")

		if presetName != "" {
			p := presets.Get(presetName)
			if p == nil {
				fmt.Println("Available presets:")
				for _, pr := range presets.All() {
					fmt.Printf("  %-20s %s\n", pr.Name, pr.Description)
				}
				return fmt.Errorf("preset %q not found", presetName)
			}
			if width == 0 {
				width = p.Width
			}
			if height == 0 {
				height = p.Height
			}
			if bg == "" {
				bg = p.Background
			}
		}

		if width == 0 {
			width = 1080
		}
		if height == 0 {
			height = 1080
		}
		if bg == "" {
			bg = "#ffffff"
		}

		proj := project.New(name, width, height, bg)
		proj.FilePath = project.ProjectFile(name)

		if err := proj.Save(); err != nil {
			return fmt.Errorf("save project: %w", err)
		}

		fmt.Printf("Created project %q (%dx%d) → %s\n", name, width, height, proj.FilePath)
		return nil
	},
}

var presetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "List available canvas presets",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available presets:")
		fmt.Println()
		for _, p := range presets.All() {
			fmt.Printf("  %-22s %s\n", p.Name, p.Description)
		}
	},
}

func init() {
	newCmd.Flags().String("preset", "", "Use a preset canvas size")
	newCmd.Flags().Int("width", 0, "Canvas width in pixels")
	newCmd.Flags().Int("height", 0, "Canvas height in pixels")
	newCmd.Flags().String("bg", "", "Background color (hex)")

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(presetsCmd)
}
