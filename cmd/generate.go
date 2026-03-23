package cmd

import (
	"fmt"

	"github.com/jwvictor/mkimg/internal/ai"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an AI image (standalone, outside a project)",
	Long: `Generate an AI image using Gemini (Nano Banana).

This is a standalone command for quick image generation without a project.
For project-based AI layers, use: mkimg layer add ai --prompt "..."

Requires GEMINI_API_KEY or GOOGLE_API_KEY environment variable.

Examples:
  mkimg generate --prompt "a sleek product photo on marble" -o product.png
  mkimg generate --prompt "abstract gradient background" --aspect 16:9 -o bg.png
  mkimg generate --prompt "similar style but darker" --reference input.png -o out.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		output, _ := cmd.Flags().GetString("output")
		aspect, _ := cmd.Flags().GetString("aspect")
		ref, _ := cmd.Flags().GetString("reference")
		model, _ := cmd.Flags().GetString("model")

		if prompt == "" {
			return fmt.Errorf("--prompt is required")
		}
		if output == "" {
			output = "generated.png"
		}

		fmt.Printf("Generating image with AI...\n")
		fmt.Printf("  Prompt: %s\n", prompt)
		if ref != "" {
			fmt.Printf("  Reference: %s\n", ref)
		}
		if aspect != "" {
			fmt.Printf("  Aspect ratio: %s\n", aspect)
		}

		err := ai.GenerateImage(prompt, output, ai.Options{
			Model:          model,
			ReferenceImage: ref,
			AspectRatio:    aspect,
		})
		if err != nil {
			return fmt.Errorf("generation failed: %w", err)
		}

		return nil
	},
}

func init() {
	generateCmd.Flags().String("prompt", "", "Image generation prompt")
	generateCmd.Flags().StringP("output", "o", "", "Output file path")
	generateCmd.Flags().String("aspect", "", "Aspect ratio (e.g., 16:9, 1:1, 9:16)")
	generateCmd.Flags().String("reference", "", "Reference image for style guidance")
	generateCmd.Flags().String("model", "", "Gemini model (default: gemini-2.5-flash-preview-image-generation)")

	rootCmd.AddCommand(generateCmd)
}
