package cmd

import (
	"fmt"

	"mkimg/internal/fonts"

	"github.com/spf13/cobra"
)

var fontCmd = &cobra.Command{
	Use:   "font",
	Short: "Manage fonts",
}

var fontSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search Google Fonts",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		limit, _ := cmd.Flags().GetInt("limit")

		results, err := fonts.SearchGoogleFonts(query)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No fonts found.")
			return nil
		}

		if limit > 0 && len(results) > limit {
			results = results[:limit]
		}

		fmt.Printf("Found %d fonts:\n\n", len(results))
		for _, f := range results {
			fmt.Printf("  %-30s %-12s %d variants\n", f.Family, f.Category, f.Variants)
		}
		return nil
	},
}

var fontInstallCmd = &cobra.Command{
	Use:   "install <family>",
	Short: "Install a font from Google Fonts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Installing font %q from Google Fonts...\n", args[0])
		return fonts.InstallFont(args[0])
	},
}

var fontListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed fonts",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		families, err := fonts.ListInstalled()
		if err != nil {
			return err
		}

		if len(families) == 0 {
			fmt.Println("No fonts installed. Install with: mkimg font install <family>")
			return nil
		}

		fmt.Printf("Installed fonts (%s):\n\n", fonts.FontDir())
		for _, f := range families {
			fmt.Printf("  %s\n", f)
		}
		return nil
	},
}

func init() {
	fontSearchCmd.Flags().Int("limit", 20, "Max results")

	fontCmd.AddCommand(fontSearchCmd)
	fontCmd.AddCommand(fontInstallCmd)
	fontCmd.AddCommand(fontListCmd)

	rootCmd.AddCommand(fontCmd)
}
