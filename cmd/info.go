package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show project info",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		fmt.Printf("Project: %s\n", proj.Name)
		fmt.Printf("File:    %s\n", proj.FilePath)
		fmt.Printf("Canvas:  %dx%d\n", proj.Canvas.Width, proj.Canvas.Height)
		fmt.Printf("Background: %s\n", proj.Canvas.Background)
		fmt.Printf("Layers:  %d\n", len(proj.Layers))
		fmt.Printf("Created: %s\n", proj.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("Updated: %s\n", proj.UpdatedAt.Format("2006-01-02 15:04"))

		return nil
	},
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump project as formatted JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(proj)
	},
}

var resizeCmd = &cobra.Command{
	Use:   "resize",
	Short: "Resize the canvas",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("width") {
			w, _ := cmd.Flags().GetInt("width")
			proj.Canvas.Width = w
		}
		if cmd.Flags().Changed("height") {
			h, _ := cmd.Flags().GetInt("height")
			proj.Canvas.Height = h
		}
		if cmd.Flags().Changed("bg") {
			bg, _ := cmd.Flags().GetString("bg")
			proj.Canvas.Background = bg
		}

		if err := proj.Save(); err != nil {
			return err
		}

		fmt.Printf("Canvas resized to %dx%d\n", proj.Canvas.Width, proj.Canvas.Height)
		return nil
	},
}

func init() {
	infoCmd.Flags().StringP("project", "p", "", "Project file")
	dumpCmd.Flags().StringP("project", "p", "", "Project file")

	resizeCmd.Flags().Int("width", 0, "New width")
	resizeCmd.Flags().Int("height", 0, "New height")
	resizeCmd.Flags().String("bg", "", "New background color")
	resizeCmd.Flags().StringP("project", "p", "", "Project file")

	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(dumpCmd)
	rootCmd.AddCommand(resizeCmd)
}
