package cmd

import (
	"fmt"

	"mkimg/internal/icons"

	"github.com/spf13/cobra"
)

var iconCmd = &cobra.Command{
	Use:   "icon",
	Short: "Manage icon libraries",
}

var iconInstallCmd = &cobra.Command{
	Use:   "install <collection>",
	Short: "Install an icon collection",
	Long: `Install an icon font collection.

Collections:
  material      Material Design Symbols (Google)
  fontawesome   Font Awesome Free

Examples:
  mkimg icon install material
  mkimg icon install fontawesome`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "material":
			fmt.Println("Installing Material Symbols...")
			return icons.InstallMaterialIcons()
		case "fontawesome":
			fmt.Println("Installing Font Awesome Free...")
			return icons.InstallFontAwesome()
		default:
			return fmt.Errorf("unknown collection %q (available: material, fontawesome)", args[0])
		}
	},
}

var iconSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for icons",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collection, _ := cmd.Flags().GetString("collection")
		limit, _ := cmd.Flags().GetInt("limit")

		if collection == "" || collection == "material" {
			results, err := icons.SearchMaterialIcons(args[0])
			if err != nil {
				if collection == "material" {
					return err
				}
				// Non-fatal if searching both
			} else {
				if limit > 0 && len(results) > limit {
					results = results[:limit]
				}
				if len(results) > 0 {
					fmt.Printf("Material Symbols (%d results):\n", len(results))
					for _, name := range results {
						fmt.Printf("  %s\n", name)
					}
				}
			}
		}

		return nil
	},
}

var iconListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed icon collections",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		collections := icons.ListCollections()
		if len(collections) == 0 {
			fmt.Println("No icon collections installed.")
			fmt.Println("Install with: mkimg icon install material")
			return
		}
		fmt.Println("Installed icon collections:")
		for _, c := range collections {
			fmt.Printf("  %s\n", c)
		}
	},
}

func init() {
	iconSearchCmd.Flags().String("collection", "", "Search specific collection")
	iconSearchCmd.Flags().Int("limit", 20, "Max results")

	iconCmd.AddCommand(iconInstallCmd)
	iconCmd.AddCommand(iconSearchCmd)
	iconCmd.AddCommand(iconListCmd)

	rootCmd.AddCommand(iconCmd)
}
