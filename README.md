# mcpautocrop

A lightweight MCP (Model Context Protocol) server that exposes an `auto_crop_image` tool to AI agents. Given an image, it detects the background color from the top-left pixel, calculates the bounding box of the subject, and saves a cropped copy.

## Features

- **Zero runtime dependencies** — single static binary
- **PNG and JPEG support** — output format matches the file extension of `output_path`
- **Fast** — pure Go pixel iteration; a 1080p image processes in under 100 ms
- **MCP stdio transport** — works with Claude Desktop and any MCP-compatible agent

## Tool: `auto_crop_image`

| Parameter | Type | Required | Description |
|---|---|---|---|
| `input_path` | string | yes | Source image file (PNG or JPEG) |
| `output_path` | string | yes | Destination file for the cropped result |
| `border` | integer | no | Pixels of padding to add around the subject (default `0`) |

**Return values:**
- `Successfully cropped and saved to <output_path>`
- `Image is already optimally cropped. No changes made.`
- `Image consists entirely of background color; cannot crop.`
- Error message on file/decode/write failure

## Getting Started

### Build

```bash
make build
```

### Run unit tests

```bash
make test
```

### Live crop demo (downloads a sample PNG and runs it through the tool)

```bash
make demo
# Downloads testdata/sample.png, crops it, saves testdata/sample_cropped.png
```

### Cross-compile release binaries

```bash
make release
# Outputs to dist/
```

## CLI Usage

The binary has two modes:

```
mcpautocrop mcp                         Run as MCP server (stdio transport)
mcpautocrop test <input> <output>       Crop an image and print the result
mcpautocrop --version | -v              Print version and exit
```

**Example:**

```bash
# Tight crop (no padding)
./mcpautocrop test photo.png photo_cropped.png

# Add 10 pixels of padding around the subject
./mcpautocrop test --border 10 photo.png photo_cropped.png
# Successfully cropped and saved to photo_cropped.png
```

> **Note:** `--border` must appear before the positional arguments.

## MCP Configuration (Claude Desktop)

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mcpautocrop": {
      "command": "/path/to/mcpautocrop",
      "args": ["mcp"]
    }
  }
}
```

## How It Works

1. The server starts and listens on **stdio** using the MCP protocol.
2. When `auto_crop_image` is called, it:
   - Opens and decodes the source image.
   - Samples the top-left pixel `(0,0)` as the background color.
   - Iterates all pixels to find the minimum bounding rectangle of non-background pixels.
   - Crops via `SubImage` and encodes to the output path.

## Make Targets

| Target | Description |
|---|---|
| `make build` | Compile the binary |
| `make test` | Run all tests with race detector |
| `make run` | Build and run the MCP server (stdio) |
| `make fetch-testimg` | Download sample PNG into `testdata/` |
| `make demo` | Download sample, build, and run a live crop |
| `make clean` | Remove build artifacts |
| `make lint` | Run `go vet` (and `golangci-lint` if available) |
| `make fmt` | Format code with `gofmt`/`goimports` |
| `make tidy` | Run `go mod tidy` |
| `make version` | Display current version |
| `make version-increment` | Bump the patch version |
| `make release` | Build release binaries for Linux, macOS, Windows |
| `make push` | Full release cycle: test → bump → commit → tag → push |

## License

See [LICENSE](LICENSE).
