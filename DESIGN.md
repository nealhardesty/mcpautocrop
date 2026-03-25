# Product Requirements Document (PRD): Auto-Crop MCP Server (Go Implementation)

## 1. Overview
The Auto-Crop MCP (Model Context Protocol) Server is a lightweight, compiled microservice designed to expose image processing capabilities to Large Language Models (LLMs). By leveraging Go's standard library and fast execution, this server provides a highly performant, single-binary tool that analyzes an image, detects the "content" bounding box, and automatically crops away unnecessary backgrounds.

## 2. Objectives
* Provide LLMs with a reliable `auto_crop` tool via a standard MCP interface.
* Utilize Go's native `image` package to eliminate the need for heavy external dependencies (like C-bindings or Python environments).
* Deliver a lightweight, instantly booting executable with a minimal memory footprint, ideal for background agent processes.

## 3. Target Audience / Use Cases
* **Primary Users:** AI Assistants and Agents (e.g., Claude Desktop, custom Go-based or Node-based LLM agents) utilizing the Model Context Protocol.
* **Use Case:** An AI agent is tasked with formatting a directory of raw screenshots or product photos. The agent calls the `auto_crop_image` tool on each file to remove redundant white space before using them in a downstream task.

---

## 4. Technical Stack
* **Language:** Go 1.21+
* **Framework:** `github.com/mark3labs/mcp-go` (or equivalent community SDK for rapid MCP server deployment and tool registration).
* **Image Processing:** Go Standard Library (`image`, `image/color`, `image/png`, `image/jpeg`).

---

## 5. Functional Requirements
### 5.1. Core Capabilities
* **Read Image:** The server must read common image formats (PNG, JPEG) from the local filesystem using `os.Open` and `image.Decode`.
* **Detect Background:** The server must identify the background color by sampling the top-left pixel `(0, 0)`.
* **Calculate Bounding Box:** The server must iterate through the image pixels to calculate the minimum bounding rectangle (`image.Rectangle`) containing pixels that *do not* match the background color.
* **Crop and Save:** The server must cast the image to a `SubImager`, crop it to the calculated bounding box, and save it to a provided destination path using `png.Encode` or `jpeg.Encode`.

### 5.2. MCP Tool Interface
The server will expose a single tool to the LLM.

**Tool Name:** `auto_crop_image`
**Description:** Analyzes an image file, calculates the bounding box of the core subject by checking against the background color, crops the image, and saves the result.

**Input Parameters:**
* `input_path` (string, required): The absolute or relative file path to the source image.
* `output_path` (string, required): The file path where the cropped image should be saved.

**Output (Return String):**
* Success message confirming the file was cropped and saved (e.g., `Successfully cropped and saved to {output_path}`).
* OR a message indicating no cropping was necessary.
* OR an error message detailing what went wrong.

---

## 6. Non-Functional Requirements
* **Performance:** Because Go compiles to machine code, iterating over pixels is highly efficient. Processing a standard 1080p image should take under 100 milliseconds.
* **Portability:** The server must be able to be cross-compiled into a single standalone executable for macOS, Linux, and Windows without requiring the user to install any runtimes.
* **Statelessness:** The server maintains no internal state between tool calls. 
* **Security:** The tool operates strictly on local files. It relies on the OS-level file permissions of the user running the MCP binary.

---

## 7. Edge Cases & Error Handling

| Scenario | Expected Server Behavior |
| :--- | :--- |
| **File Not Found** | Return a graceful error: `Error: Input file '{input_path}' could not be opened.` |
| **Unsupported/Corrupt File Type** | Return an error: `Error: image decoding failed. Ensure file is a valid PNG or JPEG.` |
| **No Cropping Needed** | Return a specific message: `Image is already optimally cropped. No changes made.` |
| **Fully Empty Image** (e.g., pure white) | Return a specific message: `Image consists entirely of background color; cannot crop.` |
| **Write Permission Denied** | Return an error: `Error: Cannot write to '{output_path}'. Permission denied.` |

---

## 8. Future Enhancements (V2)
* **Color Tolerance Parameter:** Implement a Euclidean distance check in RGB space to allow the LLM to specify a "tolerance" level (e.g., for noisy JPEGs where the background isn't perfectly uniform).
* **Padding Parameter:** Allow the LLM to request a specific amount of padding (in pixels) around the cropped subject.
* **Concurrency:** If handling massive images, implement Go routines to scan image chunks (top, bottom, left, right) concurrently to find the bounding box even faster.
