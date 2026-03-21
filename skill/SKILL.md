---
name: mkimg
version: 1.0.0
description: "CLI image editor and generator for ad creatives, social media graphics, and visual content."
metadata:
  openclaw:
    category: "creative"
    emoji: "🖼️"
    requires:
      bins: ["mkimg"]
      env: ["GEMINI_API_KEY"]
    install:
      go: "github.com/jwvictor/mkimg@latest"
---

# mkimg

A CLI image editor and generator for creating ad creatives, social media graphics, and visual content. Works through a declarative JSON project format (`.mkimg.json`) with compositable layers, filters, and AI generation.

## Core Concepts

- **Project** — A `.mkimg.json` file containing canvas dimensions, background color, and an ordered list of layers. Commands auto-detect this file in the current directory.
- **Canvas** — The output image dimensions and background color.
- **Layer** — A compositable visual element. Types: `solid`, `image`, `text`, `shape`, `gradient`, `ai`, `icon`. Each layer has an auto-generated 6-char ID, position (x/y), dimensions, opacity, rotation, visibility, and type-specific properties.
- **Filter** — A post-processing effect applied to a layer (blur, brightness, sepia, vignette, etc.).
- **Preset** — A named canvas template (e.g. `instagram-story`, `youtube-thumbnail`, `business-card`).

## Project Structure

```
mkimg/
├── main.go                     # Entry point
├── cmd/                        # CLI commands (cobra)
│   ├── root.go                 # Root command
│   ├── new.go                  # Project creation
│   ├── layer.go                # Layer CRUD (add/edit/move/remove/toggle/duplicate)
│   ├── render.go               # Render & preview
│   ├── filter.go               # Filter management
│   ├── generate.go             # Standalone AI generation
│   ├── info.go                 # Project info, dump, resize
│   ├── icon.go                 # Icon library management
│   └── font.go                 # Font management
└── internal/
    ├── project/project.go      # Project model & JSON persistence
    ├── engine/engine.go        # Rendering engine (composites layers to image)
    ├── ai/ai.go                # Gemini API for AI image generation
    ├── fonts/fonts.go          # Google Fonts search, install, loading
    ├── icons/icons.go          # Material & FontAwesome icon support
    ├── effects/effects.go      # 18 image filters/effects
    └── presets/presets.go       # 20 canvas presets (social, ads, print, web)
```

## CLI Commands

### Project Lifecycle

```bash
mkimg new <name> [--preset <name>] [--width <px>] [--height <px>] [--bg <hex>]
mkimg presets                     # List available presets
mkimg info [-p project.json]      # Show project details
mkimg dump [-p project.json]      # Print raw JSON
mkimg resize --width <px> --height <px> [--bg <hex>]
```

### Layers

```bash
mkimg layer add <type> [flags]    # Add a layer
mkimg layer list                  # List all layers
mkimg layer edit <id> [flags]     # Modify layer properties
mkimg layer move <id> <position>  # Reorder
mkimg layer remove <id>           # Delete
mkimg layer toggle <id>           # Show/hide
mkimg layer duplicate <id>        # Clone
```

**Layer types and key flags:**

| Type       | Flags                                                                 |
|------------|-----------------------------------------------------------------------|
| `solid`    | `--color`                                                             |
| `image`    | `--src`, `--fit` (cover/contain/fill/none)                            |
| `text`     | `--content`, `--font`, `--size`, `--color`, `--align`, `--max-width`, `--line-height`, `--shadow` |
| `shape`    | `--shape` (rect/circle/ellipse/line), `--fill`, `--radius`, `--stroke-color`, `--stroke-width` |
| `gradient` | `--gradient-type` (linear/radial/conic), `--from`, `--to`, `--angle`, `--stops` |
| `ai`       | `--prompt`, `--aspect`, `--reference`                                 |
| `icon`     | `--icon-name`, `--collection` (material/fontawesome), `--size`, `--color` |

**Common flags** (all layer types): `--name`, `--x`, `--y`, `--width`, `--height`, `--opacity`, `--rotation`

### Rendering

```bash
mkimg render [-o output.png] [--open]   # Render project to file
mkimg preview                            # Render and open immediately
```

Output format is determined by file extension (PNG or JPEG).

### Filters

```bash
mkimg filter <layer-id> <type> [flags]   # Apply a filter
mkimg filters                             # List all 18 filter types
mkimg unfilter <layer-id> [type]          # Remove filters
```

Available filters: `blur`, `sharpen`, `brightness`, `contrast`, `saturation`, `gamma`, `hue`, `grayscale`, `sepia`, `invert`, `duotone`, `pixelate`, `vignette`, `noise`, `posterize`, `emboss`, `edge`, `glow`

### Fonts

```bash
mkimg font search [query] [--limit <n>]  # Search Google Fonts
mkimg font install <family>               # Download font family
mkimg font list                           # Show installed fonts
```

### Icons

```bash
mkimg icon install <collection>           # Install material or fontawesome
mkimg icon search <query> [--collection] [--limit]
mkimg icon list                           # Show installed collections
```

### Standalone AI Generation

```bash
mkimg generate --prompt <text> [-o output.png] [--aspect <ratio>] [--reference <path>] [--model <model>]
```

## External Integrations

| Service | Purpose | Auth |
|---------|---------|------|
| Google Gemini API | AI image generation | `GEMINI_API_KEY` or `GOOGLE_API_KEY` env var |
| Google Fonts | Font search & download | No key needed (public metadata + CSS2 API) |
| Material Design Icons | Icon font + codepoints | No key (GitHub CDN) |
| Font Awesome | Icon fonts + metadata | No key (CDN) |

## Local Cache

```
~/.mkimg/
├── fonts/<family>/*.ttf
└── icons/
    ├── material/   (MaterialSymbolsOutlined.ttf + codepoints.txt)
    └── fontawesome/ (fa-solid-900.ttf, fa-regular-400.ttf, fa-brands-400.ttf, icons.json)
```

## Key Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/fogleman/gg` — 2D drawing/graphics context
- `github.com/disintegration/imaging` — Image transforms and filters
- `golang.org/x/image` — Image codecs and font rendering

## Example Workflow

```bash
mkimg new summer-sale --preset instagram-story --bg "#1a1a2e"
mkimg layer add gradient --from "#e94560" --to "#f5a623" --angle 160
mkimg font install Anton
mkimg layer add text --content "SUMMER SALE" --font Anton --size 120 --color "#ffffff" --align center --x 540 --y 400
mkimg layer add shape --shape circle --fill "#ffffff22" --width 300 --height 300 --x 390 --y 800
mkimg filter <gradient-id> vignette --strength 0.5
mkimg render -o summer-sale.png --open
```
