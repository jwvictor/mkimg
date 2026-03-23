package cmd

import (
	"fmt"
	"strings"

	"github.com/jwvictor/mkimg/internal/ai"
	"github.com/jwvictor/mkimg/internal/project"

	"github.com/spf13/cobra"
)

var layerCmd = &cobra.Command{
	Use:   "layer",
	Short: "Manage layers",
}

// --- layer add ---

var layerAddCmd = &cobra.Command{
	Use:   "add <type>",
	Short: "Add a layer",
	Long: `Add a layer to the project.

Types: solid, image, text, shape, gradient, ai, icon

Examples:
  mkimg layer add solid --color "#ff0000"
  mkimg layer add image --src photo.png --fit cover
  mkimg layer add text --content "Hello World" --font Montserrat --size 72
  mkimg layer add shape --shape rect --fill "#333" --width 400 --height 200 --radius 20
  mkimg layer add gradient --from "#667eea" --to "#764ba2" --angle 135
  mkimg layer add ai --prompt "minimalist product photo on white background"
  mkimg layer add icon --name heart --collection material --size 48`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		layerType := args[0]

		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		layer := project.Layer{
			Type: layerType,
		}

		// Common flags
		layer.Name, _ = cmd.Flags().GetString("name")
		layer.X, _ = cmd.Flags().GetFloat64("x")
		layer.Y, _ = cmd.Flags().GetFloat64("y")
		layer.Width, _ = cmd.Flags().GetFloat64("width")
		layer.Height, _ = cmd.Flags().GetFloat64("height")
		layer.Opacity, _ = cmd.Flags().GetFloat64("opacity")
		layer.Rotation, _ = cmd.Flags().GetFloat64("rotation")

		switch layerType {
		case "solid":
			layer.Color, _ = cmd.Flags().GetString("color")
			if layer.Color == "" {
				return fmt.Errorf("--color required for solid layer")
			}

		case "image":
			layer.Source, _ = cmd.Flags().GetString("src")
			if layer.Source == "" {
				return fmt.Errorf("--src required for image layer")
			}
			layer.Fit, _ = cmd.Flags().GetString("fit")
			layer.CropX, _ = cmd.Flags().GetFloat64("crop-x")
			layer.CropY, _ = cmd.Flags().GetFloat64("crop-y")
			layer.CropWidth, _ = cmd.Flags().GetFloat64("crop-width")
			layer.CropHeight, _ = cmd.Flags().GetFloat64("crop-height")

		case "text":
			layer.Content, _ = cmd.Flags().GetString("content")
			if layer.Content == "" {
				return fmt.Errorf("--content required for text layer")
			}
			layer.Font, _ = cmd.Flags().GetString("font")
			layer.FontSize, _ = cmd.Flags().GetFloat64("size")
			layer.FontWeight, _ = cmd.Flags().GetString("weight")
			layer.Color, _ = cmd.Flags().GetString("color")
			layer.Align, _ = cmd.Flags().GetString("align")
			layer.MaxWidth, _ = cmd.Flags().GetFloat64("max-width")
			layer.LineHeight, _ = cmd.Flags().GetFloat64("line-height")

			// Shadow shorthand
			shadowColor, _ := cmd.Flags().GetString("shadow")
			if shadowColor != "" {
				layer.ShadowEffect = &project.Shadow{
					Color:   shadowColor,
					OffsetX: 2,
					OffsetY: 2,
					Blur:    4,
				}
			}

		case "shape":
			layer.Shape, _ = cmd.Flags().GetString("shape")
			if layer.Shape == "" {
				layer.Shape = "rect"
			}
			layer.Fill, _ = cmd.Flags().GetString("fill")
			layer.Radius, _ = cmd.Flags().GetFloat64("radius")

			strokeColor, _ := cmd.Flags().GetString("stroke-color")
			strokeWidth, _ := cmd.Flags().GetFloat64("stroke-width")
			if strokeColor != "" {
				layer.StrokeStyle = &project.Stroke{
					Color: strokeColor,
					Width: strokeWidth,
				}
			}

		case "gradient":
			layer.GradientType, _ = cmd.Flags().GetString("gradient-type")
			if layer.GradientType == "" {
				layer.GradientType = "linear"
			}
			layer.GradientAngle, _ = cmd.Flags().GetFloat64("angle")

			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")
			stopsStr, _ := cmd.Flags().GetString("stops")

			if stopsStr != "" {
				// Parse stops: "color1:pos1,color2:pos2,..."
				for _, s := range strings.Split(stopsStr, ",") {
					parts := strings.SplitN(strings.TrimSpace(s), ":", 2)
					if len(parts) == 2 {
						var pos float64
						fmt.Sscanf(parts[1], "%f", &pos)
						layer.GradientStops = append(layer.GradientStops, project.GradientStop{
							Color:    parts[0],
							Position: pos,
						})
					}
				}
			} else if from != "" && to != "" {
				layer.GradientStops = []project.GradientStop{
					{Color: from, Position: 0},
					{Color: to, Position: 1},
				}
			} else {
				return fmt.Errorf("gradient requires --from/--to or --stops")
			}

		case "ai":
			layer.AIPrompt, _ = cmd.Flags().GetString("prompt")
			if layer.AIPrompt == "" {
				return fmt.Errorf("--prompt required for AI layer")
			}
			layer.AspectRatio, _ = cmd.Flags().GetString("aspect")
			ref, _ := cmd.Flags().GetString("reference")

			// Generate the image
			outputFile := fmt.Sprintf("ai_%s.png", project.GenerateID())
			fmt.Printf("Generating AI image...\n")
			err := ai.GenerateImage(layer.AIPrompt, outputFile, ai.Options{
				AspectRatio:    layer.AspectRatio,
				ReferenceImage: ref,
			})
			if err != nil {
				return fmt.Errorf("AI generation failed: %w", err)
			}
			layer.Source = outputFile

		case "icon":
			layer.IconName, _ = cmd.Flags().GetString("icon-name")
			if layer.IconName == "" {
				// Also try the shorthand --name flag
				layer.IconName, _ = cmd.Flags().GetString("name")
			}
			if layer.IconName == "" {
				return fmt.Errorf("--icon-name required for icon layer")
			}
			layer.IconCollection, _ = cmd.Flags().GetString("collection")
			layer.FontSize, _ = cmd.Flags().GetFloat64("size")
			layer.Color, _ = cmd.Flags().GetString("color")

		default:
			return fmt.Errorf("unknown layer type: %s (valid: solid, image, text, shape, gradient, ai, icon)", layerType)
		}

		if layer.Name == "" {
			layer.Name = fmt.Sprintf("%s layer", layerType)
		}

		id := proj.AddLayer(layer)
		if err := proj.Save(); err != nil {
			return err
		}

		fmt.Printf("Added %s layer %q (id: %s)\n", layerType, layer.Name, id)
		return nil
	},
}

// --- layer list ---

var layerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all layers",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		if len(proj.Layers) == 0 {
			fmt.Println("No layers. Add one with: mkimg layer add <type>")
			return nil
		}

		fmt.Printf("Project: %s (%dx%d)\n", proj.Name, proj.Canvas.Width, proj.Canvas.Height)
		fmt.Println()

		for i, l := range proj.Layers {
			vis := "+"
			if !l.Visible {
				vis = "-"
			}
			name := l.Name
			if name == "" {
				name = l.Type
			}
			extra := layerSummary(&l)
			fmt.Printf("  %d. [%s] %-8s %-6s %s %s\n", i, vis, l.ID, l.Type, name, extra)
		}
		return nil
	},
}

// --- layer remove ---

var layerRemoveCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove a layer",
	Aliases: []string{"rm"},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		if err := proj.RemoveLayer(args[0]); err != nil {
			return err
		}

		if err := proj.Save(); err != nil {
			return err
		}

		fmt.Printf("Removed layer %s\n", args[0])
		return nil
	},
}

// --- layer move ---

var layerMoveCmd = &cobra.Command{
	Use:   "move <id> <position>",
	Short: "Move a layer to a new position",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		var pos int
		fmt.Sscanf(args[1], "%d", &pos)

		if err := proj.MoveLayer(args[0], pos); err != nil {
			return err
		}

		if err := proj.Save(); err != nil {
			return err
		}

		fmt.Printf("Moved layer %s to position %d\n", args[0], pos)
		return nil
	},
}

// --- layer edit ---

var layerEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit layer properties",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		layer := proj.GetLayer(args[0])
		if layer == nil {
			return fmt.Errorf("layer %q not found", args[0])
		}

		// Update any flags that were explicitly set
		if cmd.Flags().Changed("x") {
			layer.X, _ = cmd.Flags().GetFloat64("x")
		}
		if cmd.Flags().Changed("y") {
			layer.Y, _ = cmd.Flags().GetFloat64("y")
		}
		if cmd.Flags().Changed("width") {
			layer.Width, _ = cmd.Flags().GetFloat64("width")
		}
		if cmd.Flags().Changed("height") {
			layer.Height, _ = cmd.Flags().GetFloat64("height")
		}
		if cmd.Flags().Changed("opacity") {
			layer.Opacity, _ = cmd.Flags().GetFloat64("opacity")
		}
		if cmd.Flags().Changed("rotation") {
			layer.Rotation, _ = cmd.Flags().GetFloat64("rotation")
		}
		if cmd.Flags().Changed("visible") {
			layer.Visible, _ = cmd.Flags().GetBool("visible")
		}
		if cmd.Flags().Changed("name") {
			layer.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("color") {
			layer.Color, _ = cmd.Flags().GetString("color")
		}
		if cmd.Flags().Changed("content") {
			layer.Content, _ = cmd.Flags().GetString("content")
		}
		if cmd.Flags().Changed("font") {
			layer.Font, _ = cmd.Flags().GetString("font")
		}
		if cmd.Flags().Changed("size") {
			layer.FontSize, _ = cmd.Flags().GetFloat64("size")
		}
		if cmd.Flags().Changed("src") {
			layer.Source, _ = cmd.Flags().GetString("src")
		}
		if cmd.Flags().Changed("crop-x") {
			layer.CropX, _ = cmd.Flags().GetFloat64("crop-x")
		}
		if cmd.Flags().Changed("crop-y") {
			layer.CropY, _ = cmd.Flags().GetFloat64("crop-y")
		}
		if cmd.Flags().Changed("crop-width") {
			layer.CropWidth, _ = cmd.Flags().GetFloat64("crop-width")
		}
		if cmd.Flags().Changed("crop-height") {
			layer.CropHeight, _ = cmd.Flags().GetFloat64("crop-height")
		}
		if cmd.Flags().Changed("fill") {
			layer.Fill, _ = cmd.Flags().GetString("fill")
		}

		if err := proj.Save(); err != nil {
			return err
		}

		fmt.Printf("Updated layer %s\n", args[0])
		return nil
	},
}

// --- layer toggle ---

var layerToggleCmd = &cobra.Command{
	Use:   "toggle <id>",
	Short: "Toggle layer visibility",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		layer := proj.GetLayer(args[0])
		if layer == nil {
			return fmt.Errorf("layer %q not found", args[0])
		}

		layer.Visible = !layer.Visible
		if err := proj.Save(); err != nil {
			return err
		}

		state := "visible"
		if !layer.Visible {
			state = "hidden"
		}
		fmt.Printf("Layer %s is now %s\n", args[0], state)
		return nil
	},
}

// --- layer duplicate ---

var layerDuplicateCmd = &cobra.Command{
	Use:   "duplicate <id>",
	Short: "Duplicate a layer",
	Aliases: []string{"dup"},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := loadProject(cmd)
		if err != nil {
			return err
		}

		layer := proj.GetLayer(args[0])
		if layer == nil {
			return fmt.Errorf("layer %q not found", args[0])
		}

		newLayer := *layer
		newLayer.ID = project.GenerateID()
		newLayer.Name = layer.Name + " (copy)"

		proj.AddLayer(newLayer)
		if err := proj.Save(); err != nil {
			return err
		}

		fmt.Printf("Duplicated layer %s → %s\n", args[0], newLayer.ID)
		return nil
	},
}

func init() {
	// Add subcommands
	layerCmd.AddCommand(layerAddCmd)
	layerCmd.AddCommand(layerListCmd)
	layerCmd.AddCommand(layerRemoveCmd)
	layerCmd.AddCommand(layerMoveCmd)
	layerCmd.AddCommand(layerEditCmd)
	layerCmd.AddCommand(layerToggleCmd)
	layerCmd.AddCommand(layerDuplicateCmd)

	// Common layer flags
	addLayerFlags := layerAddCmd.Flags()
	addLayerFlags.String("name", "", "Layer name")
	addLayerFlags.Float64("x", 0, "X position")
	addLayerFlags.Float64("y", 0, "Y position")
	addLayerFlags.Float64("width", 0, "Width")
	addLayerFlags.Float64("height", 0, "Height")
	addLayerFlags.Float64("opacity", 1.0, "Opacity (0-1)")
	addLayerFlags.Float64("rotation", 0, "Rotation in degrees")

	// Type-specific flags
	addLayerFlags.String("color", "", "Color (hex)")
	addLayerFlags.String("src", "", "Image source path")
	addLayerFlags.String("fit", "", "Image fit: cover, contain, fill, none")
	addLayerFlags.Float64("crop-x", 0, "Crop region X offset in source image")
	addLayerFlags.Float64("crop-y", 0, "Crop region Y offset in source image")
	addLayerFlags.Float64("crop-width", 0, "Crop region width in source image")
	addLayerFlags.Float64("crop-height", 0, "Crop region height in source image")
	addLayerFlags.String("content", "", "Text content")
	addLayerFlags.String("font", "", "Font family")
	addLayerFlags.Float64("size", 0, "Font size / icon size")
	addLayerFlags.String("weight", "", "Font weight")
	addLayerFlags.String("align", "", "Text alignment: left, center, right")
	addLayerFlags.Float64("max-width", 0, "Max text width for wrapping")
	addLayerFlags.Float64("line-height", 0, "Line height multiplier")
	addLayerFlags.String("shadow", "", "Shadow color (hex, shorthand)")
	addLayerFlags.String("shape", "", "Shape type: rect, circle, ellipse, line")
	addLayerFlags.String("fill", "", "Shape fill color")
	addLayerFlags.Float64("radius", 0, "Corner radius")
	addLayerFlags.String("stroke-color", "", "Stroke color")
	addLayerFlags.Float64("stroke-width", 0, "Stroke width")
	addLayerFlags.String("gradient-type", "", "Gradient type: linear, radial, conic")
	addLayerFlags.String("from", "", "Gradient start color")
	addLayerFlags.String("to", "", "Gradient end color")
	addLayerFlags.Float64("angle", 0, "Gradient angle (degrees)")
	addLayerFlags.String("stops", "", "Gradient stops: '#color1:0,#color2:0.5,#color3:1'")
	addLayerFlags.String("prompt", "", "AI image prompt")
	addLayerFlags.String("aspect", "", "AI aspect ratio (e.g., 16:9)")
	addLayerFlags.String("reference", "", "AI reference image path")
	addLayerFlags.String("icon-name", "", "Icon name")
	addLayerFlags.String("collection", "", "Icon collection: material, fontawesome")

	// Edit flags
	editFlags := layerEditCmd.Flags()
	editFlags.String("name", "", "Layer name")
	editFlags.Float64("x", 0, "X position")
	editFlags.Float64("y", 0, "Y position")
	editFlags.Float64("width", 0, "Width")
	editFlags.Float64("height", 0, "Height")
	editFlags.Float64("opacity", 1.0, "Opacity (0-1)")
	editFlags.Float64("rotation", 0, "Rotation")
	editFlags.Bool("visible", true, "Visible")
	editFlags.String("color", "", "Color")
	editFlags.String("content", "", "Text content")
	editFlags.String("font", "", "Font family")
	editFlags.Float64("size", 0, "Font size")
	editFlags.String("src", "", "Image source")
	editFlags.Float64("crop-x", 0, "Crop region X offset")
	editFlags.Float64("crop-y", 0, "Crop region Y offset")
	editFlags.Float64("crop-width", 0, "Crop region width")
	editFlags.Float64("crop-height", 0, "Crop region height")
	editFlags.String("fill", "", "Fill color")

	rootCmd.AddCommand(layerCmd)
}

func loadProject(cmd *cobra.Command) (*project.Project, error) {
	path, _ := cmd.Root().Flags().GetString("project")
	if path != "" {
		return project.Load(path)
	}
	return project.FindProject()
}

func layerSummary(l *project.Layer) string {
	switch l.Type {
	case "solid":
		return l.Color
	case "image", "ai":
		return l.Source
	case "text":
		s := l.Content
		if len(s) > 30 {
			s = s[:27] + "..."
		}
		return fmt.Sprintf("%q", s)
	case "shape":
		return fmt.Sprintf("%s %s", l.Shape, l.Fill)
	case "gradient":
		if len(l.GradientStops) >= 2 {
			return fmt.Sprintf("%s %s→%s", l.GradientType, l.GradientStops[0].Color, l.GradientStops[len(l.GradientStops)-1].Color)
		}
		return l.GradientType
	case "icon":
		return fmt.Sprintf("%s/%s", l.IconCollection, l.IconName)
	default:
		return ""
	}
}
