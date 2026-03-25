package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const usage = `mcpautocrop %s

Usage:
  mcpautocrop mcp                                   Run as an MCP server (stdio transport)
  mcpautocrop test [--border N] <input> <output>    Crop an image and print the result
  mcpautocrop --version | -v                        Print version and exit

Flags for 'test':
  --border N    Add N pixels of padding around the detected subject (default 0)

MCP Configuration:
  Add mcpautocrop to Claude Desktop (~/.config/claude/claude_desktop_config.json)
  or Cursor (~/.cursor/mcp.json):

  {
    "mcpServers": {
      "mcpautocrop": {
        "command": "%s",
        "args": ["mcp"]
      }
    }
  }
`

func main() {
	exe, _ := os.Executable()
	printUsage := func() {
		fmt.Fprintf(os.Stderr, usage, Version, exe)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--version", "-v":
		fmt.Printf("mcpautocrop version %s\n", Version)

	case "mcp":
		runMCP()

	case "test":
		runTest()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// runMCP starts the MCP stdio server.
func runMCP() {
	s := server.NewMCPServer(
		"mcpautocrop",
		Version,
		server.WithToolCapabilities(true),
	)

	autoCropTool := mcp.NewTool(
		"auto_crop_image",
		mcp.WithDescription(
			"Analyzes an image file, calculates the bounding box of the core subject by "+
				"checking against the background color (sampled from the top-left pixel), "+
				"crops the image, and saves the result to the output path.",
		),
		mcp.WithString(
			"input_path",
			mcp.Required(),
			mcp.Description("The absolute or relative file path to the source image (PNG or JPEG)."),
		),
		mcp.WithString(
			"output_path",
			mcp.Required(),
			mcp.Description("The file path where the cropped image should be saved."),
		),
		mcp.WithNumber(
			"border",
			mcp.Description("Pixels of padding to add around the detected subject (default 0)."),
		),
	)

	s.AddTool(autoCropTool, autoCropHandler)

	if err := server.NewStdioServer(s).Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

// runTest runs AutoCrop from the command line and prints the result.
func runTest() {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	border := fs.Int("border", 0, "pixels of padding to add around the detected subject")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: mcpautocrop test [--border N] <input> <output>\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}
	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(1)
	}
	input, output := fs.Arg(0), fs.Arg(1)

	msg, err := AutoCrop(input, output, *border)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(msg)
}

// autoCropHandler is the MCP tool handler for auto_crop_image.
func autoCropHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	inputPath, err := req.RequireString("input_path")
	if err != nil {
		return mcp.NewToolResultError("input_path is required"), nil
	}
	outputPath, err := req.RequireString("output_path")
	if err != nil {
		return mcp.NewToolResultError("output_path is required"), nil
	}

	border := mcp.ParseInt(req, "border", 0)

	msg, err := AutoCrop(inputPath, outputPath, border)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(msg), nil
}
