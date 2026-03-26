# Changelog

## [0.1.3] - 2026-03-25

### Added
- `tolerance` parameter for `AutoCrop`: max Euclidean RGB distance for a pixel to be treated as background (default 10). Fixes incorrect "already optimally cropped" result on images with anti-aliased or slightly non-uniform white borders (e.g. real-world PNG screenshots).
- `--tolerance N` flag for the `test` CLI subcommand (default 10)
- `tolerance` optional integer parameter for the `auto_crop_image` MCP tool (default 10)
- `colorDistance()` helper using Euclidean distance in 8-bit RGB space
- `TestAutoCrop_Tolerance` test verifying exact-match failure and tolerance-based success

### Fixed
- Images with near-white (non-uniform) borders — such as browser screenshots — were incorrectly reported as "already optimally cropped" due to strict exact-match pixel comparison

## [0.1.2] - 2026-03-25

### Added
- `border` parameter for `AutoCrop`: adds N pixels of padding around the detected subject, clamped to image edges (default 0, tight crop)
- `--border N` flag for the `test` CLI subcommand
- `border` optional integer parameter for the `auto_crop_image` MCP tool
- Two new tests: `TestAutoCrop_Border` (verifies exact padded dimensions) and `TestAutoCrop_BorderClampedToImageEdge` (verifies clamping to image bounds)

## [0.1.1] - 2026-03-25

### Added
- `test` CLI subcommand: `mcpautocrop test <input> <output>` for direct command-line use
- `mcp` CLI subcommand: `mcpautocrop mcp` to start the MCP server (old default behaviour now explicit)
- Usage/help text printed when invoked with no arguments or an unknown command
- `make fetch-testimg` Makefile target: downloads `testdata/sample.png` from the Go image repository
- `make demo` Makefile target: downloads sample, builds, and runs a live crop end-to-end
- Updated `claude_desktop_config.json` example to include `"args": ["mcp"]`

## [0.1.0] - 2026-03-25

### Added
- Initial implementation of the `auto_crop_image` MCP tool
- Background detection via top-left pixel sampling
- Bounding box calculation with full pixel iteration
- PNG and JPEG output support (format chosen by output file extension)
- MCP stdio server using `github.com/mark3labs/mcp-go`
- `--version` / `-v` CLI flag
- Unit tests covering: normal crop, already-optimal, entirely-background, file-not-found, corrupt file, single-pixel subject
- `Makefile` with build, test, run, clean, lint, fmt, tidy, version, release, push targets
- `version.go` with semantic versioning
- `CHANGELOG.md` and updated `README.md`
