package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: b64 [OPTIONS] [FILE|URL]\n\n")
		fmt.Fprintf(os.Stderr, "Extract base64 encoded images from text or JSON to decoded/ directory.\n")
		fmt.Fprintf(os.Stderr, "Or encode image files to base64 format.\n")
		fmt.Fprintf(os.Stderr, "Or download images from URL and encode to base64 format.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  FILE|URL              Input file or URL to process (reads from stdin if not provided)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -f, --format-json     Pretty print JSON output (JSON input only)\n")
		fmt.Fprintf(os.Stderr, "  -p, --pretty          Pretty print JSON output (JSON input only)\n")
		fmt.Fprintf(os.Stderr, "  -o, --output DIR      Output directory for encoded image files (image input only)\n")
		fmt.Fprintf(os.Stderr, "  -h, --help            Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "Supported Formats:\n")
		fmt.Fprintf(os.Stderr, "  - JSON files with base64 images (will be parsed and formatted)\n")
		fmt.Fprintf(os.Stderr, "  - Plain text with data URLs (e.g., data:image/png;base64,...)\n")
		fmt.Fprintf(os.Stderr, "  - Markdown with embedded images (e.g., ![alt](data:image/...))\n")
		fmt.Fprintf(os.Stderr, "  - Image files (PNG, JPEG, GIF, WebP, BMP, SVG)\n")
		fmt.Fprintf(os.Stderr, "  - HTTP/HTTPS URLs pointing to image files\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  b64 s.json | jq                # Process JSON file, output compact JSON\n")
		fmt.Fprintf(os.Stderr, "  b64 --pretty s.json            # Process JSON file, output pretty JSON\n")
		fmt.Fprintf(os.Stderr, "  b64 image.png                  # Encode image to base64 (same directory)\n")
		fmt.Fprintf(os.Stderr, "  b64 -o ./ image.png            # Encode image to base64 (current directory)\n")
		fmt.Fprintf(os.Stderr, "  b64 -o /tmp image.png          # Encode image to base64 (specified directory)\n")
		fmt.Fprintf(os.Stderr, "  b64 document.md                # Process markdown/text file\n")
		fmt.Fprintf(os.Stderr, "  b64 http://example.com/pic.jpg # Download and encode image from URL\n")
		fmt.Fprintf(os.Stderr, "  cat s.json | b64 | jq          # Process from stdin\n")
		fmt.Fprintf(os.Stderr, "  cat s.json | b64 -f | jq       # Process from stdin with pretty output\n")
	}

	// 定义命令行参数
	var pretty bool
	var outputDir string
	flag.BoolVar(&pretty, "pretty", false, "pretty print JSON output")
	flag.BoolVar(&pretty, "p", false, "pretty print JSON output")
	flag.BoolVar(&pretty, "format-json", false, "pretty print JSON output")
	flag.BoolVar(&pretty, "f", false, "pretty print JSON output")
	flag.StringVar(&outputDir, "output", "", "output directory for encoded image files")
	flag.StringVar(&outputDir, "o", "", "output directory for encoded image files")
	flag.Parse()

	var data []byte
	var err error

	// 获取非标志参数（文件名或 URL）
	args := flag.Args()
	if len(args) > 0 {
		input := args[0]

		// 检查是否是 URL
		if isURL(input) {
			// 处理 URL 输入
			if err := processURLInput(input, outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing URL: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// 检查是否是图片文件
		if isImageFile(input) {
			// 处理图片文件，生成 base64 文件
			if err := processImageFile(input, outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing image file: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// 检查是否是 base64 编码文件（.mime.b64 或 .raw.b64）
		if isBase64File(input) {
			// 处理 base64 文件，解码为图片
			if err := decodeBase64File(input, outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error decoding base64 file: %v\n", err)
				os.Exit(1)
			}
			return
		}

		data, err = os.ReadFile(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", input, err)
			os.Exit(1)
		}
	} else {
		// 从标准输入读取
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
	}

	// 尝试解析为 JSON
	var result interface{}
	if err := json.Unmarshal(data, &result); err == nil {
		// 成功解析为 JSON

		// 处理 base64 图片
		if err := processImages(result, outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing images: %v\n", err)
			os.Exit(1)
		}

		// 输出处理后的 JSON
		var output []byte
		if pretty {
			output, err = json.MarshalIndent(result, "", "  ")
		} else {
			output, err = json.Marshal(result)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	} else {
		// 不是 JSON，作为纯文本处理
		if pretty {
			fmt.Fprintf(os.Stderr, "Warning: --pretty flag only applies to JSON input, ignoring\n")
		}
		text := string(data)
		processedText := processTextContent(text, outputDir)
		fmt.Print(processedText)
	}
}
