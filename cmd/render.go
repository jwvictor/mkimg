package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/jwvictor/mkimg/internal/engine"

	"github.com/spf13/cobra"
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render project to an image file",
	Long: `Render the project to PNG or JPEG.

Examples:
  mkimg render
  mkimg render -o banner.png
  mkimg render -o ad.jpg
  mkimg render --open`,
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = proj.Name + ".png"
		}

		fmt.Printf("Rendering %s (%dx%d, %d layers)...\n", proj.Name, proj.Canvas.Width, proj.Canvas.Height, len(proj.Layers))

		if err := engine.RenderToFile(proj, output); err != nil {
			return fmt.Errorf("render failed: %w", err)
		}

		fmt.Printf("Saved to %s\n", output)

		shouldOpen, _ := cmd.Flags().GetBool("open")
		if shouldOpen {
			openFile(output)
		}

		return nil
	},
}

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Render and open the image",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		output := proj.Name + "_preview.png"

		fmt.Printf("Rendering preview...\n")
		if err := engine.RenderToFile(proj, output); err != nil {
			return fmt.Errorf("render failed: %w", err)
		}

		fmt.Printf("Opening %s\n", output)
		openFile(output)
		return nil
	},
}

func openFile(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		fmt.Printf("Open %s manually\n", path)
		return
	}
	cmd.Start()
}

func init() {
	renderCmd.Flags().StringP("output", "o", "", "Output file path (default: <project>.png)")
	renderCmd.Flags().Bool("open", false, "Open the image after rendering")

	rootCmd.AddCommand(renderCmd)
	rootCmd.AddCommand(previewCmd)
}
