package cmd

import (
	"fmt"
	"strings"

	"mkimg/internal/effects"
	"mkimg/internal/project"

	"github.com/spf13/cobra"
)

var filterCmd = &cobra.Command{
	Use:   "filter <layer-id> <filter-type>",
	Short: "Apply a filter/effect to a layer",
	Long: `Apply visual filters and effects to a layer.

Available filters:
  blur         Gaussian blur (--radius)
  sharpen      Sharpen (--radius)
  brightness   Adjust brightness (--value: -100 to 100)
  contrast     Adjust contrast (--value: -100 to 100)
  saturation   Adjust saturation (--value: -100 to 100)
  gamma        Adjust gamma (--value)
  hue          Rotate hue (--value: -180 to 180)
  grayscale    Convert to grayscale
  sepia        Apply sepia tone
  invert       Invert colors
  pixelate     Pixelate (--size)
  vignette     Vignette effect (--strength: 0 to 1)
  noise        Add noise (--amount)
  posterize    Reduce color levels (--levels)
  emboss       Emboss effect
  edge         Edge detection
  glow         Soft glow (--radius, --strength)
  duotone      Two-tone (--r1,--g1,--b1,--r2,--g2,--b2)

Examples:
  mkimg filter abc123 blur --radius 5
  mkimg filter abc123 brightness --value 20
  mkimg filter abc123 sepia
  mkimg filter abc123 vignette --strength 0.7`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		layerID := args[0]
		filterType := args[1]

		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		layer := proj.GetLayer(layerID)
		if layer == nil {
			return fmt.Errorf("layer %q not found", layerID)
		}

		// Build params from flags
		params := map[string]float64{}
		for _, name := range []string{"radius", "value", "size", "strength", "amount", "levels",
			"r1", "g1", "b1", "r2", "g2", "b2"} {
			if cmd.Flags().Changed(name) {
				v, _ := cmd.Flags().GetFloat64(name)
				params[name] = v
			}
		}

		filter := project.Filter{
			Type:   filterType,
			Params: params,
		}

		layer.Filters = append(layer.Filters, filter)

		if err := proj.Save(); err != nil {
			return err
		}

		fmt.Printf("Applied %s filter to layer %s\n", filterType, layerID)
		return nil
	},
}

var filterListCmd = &cobra.Command{
	Use:   "filters",
	Short: "List available filters",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available filters:")
		fmt.Println()
		for _, f := range effects.ListFilters() {
			fmt.Printf("  %-12s %s\n", f.Name, f.Description)
			fmt.Printf("  %s  params: %s\n", strings.Repeat(" ", 12), f.Params)
		}
	},
}

var filterRemoveCmd = &cobra.Command{
	Use:   "unfilter <layer-id> [filter-type]",
	Short: "Remove filters from a layer",
	Long: `Remove filters from a layer. If filter-type is specified, only that type is removed.
If omitted, all filters are removed.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		layer := proj.GetLayer(args[0])
		if layer == nil {
			return fmt.Errorf("layer %q not found", args[0])
		}

		if len(args) == 1 {
			layer.Filters = nil
			fmt.Printf("Removed all filters from layer %s\n", args[0])
		} else {
			filterType := args[1]
			var remaining []project.Filter
			removed := 0
			for _, f := range layer.Filters {
				if f.Type == filterType {
					removed++
				} else {
					remaining = append(remaining, f)
				}
			}
			layer.Filters = remaining
			fmt.Printf("Removed %d %s filter(s) from layer %s\n", removed, filterType, args[0])
		}

		return proj.Save()
	},
}

func init() {
	filterCmd.Flags().Float64("radius", 0, "Blur/sharpen radius")
	filterCmd.Flags().Float64("value", 0, "Adjustment value")
	filterCmd.Flags().Float64("size", 0, "Pixelate block size")
	filterCmd.Flags().Float64("strength", 0, "Effect strength")
	filterCmd.Flags().Float64("amount", 0, "Noise amount")
	filterCmd.Flags().Float64("levels", 0, "Posterize levels")
	filterCmd.Flags().Float64("r1", 0, "Duotone shadow red")
	filterCmd.Flags().Float64("g1", 0, "Duotone shadow green")
	filterCmd.Flags().Float64("b1", 0, "Duotone shadow blue")
	filterCmd.Flags().Float64("r2", 255, "Duotone highlight red")
	filterCmd.Flags().Float64("g2", 200, "Duotone highlight green")
	filterCmd.Flags().Float64("b2", 100, "Duotone highlight blue")
	filterCmd.Flags().StringP("project", "p", "", "Project file")

	filterRemoveCmd.Flags().StringP("project", "p", "", "Project file")

	rootCmd.AddCommand(filterCmd)
	rootCmd.AddCommand(filterListCmd)
	rootCmd.AddCommand(filterRemoveCmd)
}
