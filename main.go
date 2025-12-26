package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

var imageCounter uint64

func main() {
	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: b64 [OPTIONS] [FILE]\n\n")
		fmt.Fprintf(os.Stderr, "Extract base64 encoded images from text or JSON to decoded/ directory.\n")
		fmt.Fprintf(os.Stderr, "Or encode image files to base64 format.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  FILE                  Input file to process (reads from stdin if not provided)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -f, --format-json     Pretty print JSON output (JSON input only)\n")
		fmt.Fprintf(os.Stderr, "  -p, --pretty          Pretty print JSON output (JSON input only)\n")
		fmt.Fprintf(os.Stderr, "  -o, --output DIR      Output directory for encoded image files (image input only)\n")
		fmt.Fprintf(os.Stderr, "  -h, --help            Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "Supported Formats:\n")
		fmt.Fprintf(os.Stderr, "  - JSON files with base64 images (will be parsed and formatted)\n")
		fmt.Fprintf(os.Stderr, "  - Plain text with data URLs (e.g., data:image/png;base64,...)\n")
		fmt.Fprintf(os.Stderr, "  - Markdown with embedded images (e.g., ![alt](data:image/...))\n")
		fmt.Fprintf(os.Stderr, "  - Image files (PNG, JPEG, GIF, WebP, BMP, SVG)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  b64 s.json | jq                # Process JSON file, output compact JSON\n")
		fmt.Fprintf(os.Stderr, "  b64 --pretty s.json            # Process JSON file, output pretty JSON\n")
		fmt.Fprintf(os.Stderr, "  b64 image.png                  # Encode image to base64 (same directory)\n")
		fmt.Fprintf(os.Stderr, "  b64 -o ./ image.png            # Encode image to base64 (current directory)\n")
		fmt.Fprintf(os.Stderr, "  b64 -o /tmp image.png          # Encode image to base64 (specified directory)\n")
		fmt.Fprintf(os.Stderr, "  b64 document.md                # Process markdown/text file\n")
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

	// 获取非标志参数（文件名）
	args := flag.Args()
	if len(args) > 0 {
		// 从文件读取
		filename := args[0]

		// 检查是否是图片文件
		if isImageFile(filename) {
			// 处理图片文件，生成 base64 文件
			if err := processImageFile(filename, outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing image file: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// 检查是否是 base64 编码文件（.mime.b64 或 .raw.b64）
		if isBase64File(filename) {
			// 处理 base64 文件，解码为图片
			if err := decodeBase64File(filename, outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error decoding base64 file: %v\n", err)
				os.Exit(1)
			}
			return
		}

		data, err = os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
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

// processTextContent 处理纯文本内容，查找并替换 base64 图片
func processTextContent(text, outputDir string) string {
	// 处理 Markdown 格式: ![alt](data:image/png;base64,...)
	mdRe := regexp.MustCompile(`!\[([^\]]*)\]\(data:(image/[^;]+);base64,([^)]+)\)`)
	text = mdRe.ReplaceAllStringFunc(text, func(match string) string {
		matches := mdRe.FindStringSubmatch(match)
		if len(matches) == 4 {
			altText := matches[1]
			mimeType := matches[2]
			base64Data := matches[3]

			filename, err := saveBase64Image(base64Data, mimeType, outputDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save markdown image: %v\n", err)
				return match // 保持原样
			}
			return fmt.Sprintf("![%s](%s)", altText, filename)
		}
		return match
	})

	// 处理普通 Data URL 格式: data:image/png;base64,...
	dataURLRe := regexp.MustCompile(`data:(image/[^;]+);base64,([A-Za-z0-9+/=]+)`)
	text = dataURLRe.ReplaceAllStringFunc(text, func(match string) string {
		matches := dataURLRe.FindStringSubmatch(match)
		if len(matches) == 3 {
			mimeType := matches[1]
			base64Data := matches[2]

			filename, err := saveBase64Image(base64Data, mimeType, outputDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save data URL image: %v\n", err)
				return match // 保持原样
			}
			return filename
		}
		return match
	})

	return text
}

// processImages 递归处理 JSON 数据，查找并保存 base64 图片
func processImages(data interface{}, outputDir string) error {
	switch v := data.(type) {
	case map[string]interface{}:
		// 检查是否包含图片数据（原格式：mime_type + data 字段）
		if mimeType, ok := v["mime_type"].(string); ok {
			if strings.HasPrefix(mimeType, "image/") {
				if dataStr, ok := v["data"].(string); ok {
					// 保存图片并替换数据
					filename, err := saveBase64Image(dataStr, mimeType, outputDir)
					if err != nil {
						return err
					}
					v["data"] = filename
				}
			}
		}

		// 递归处理所有字段，同时检查 Data URL 格式
		for key, value := range v {
			// 检查字符串值是否是 Data URL 格式
			if strValue, ok := value.(string); ok {
				if filename, replaced := processDataURL(strValue, outputDir); replaced {
					v[key] = filename
					continue
				}
			}

			// 递归处理嵌套结构
			if err := processImages(value, outputDir); err != nil {
				return err
			}
		}

	case []interface{}:
		// 递归处理数组
		for i, item := range v {
			// 检查数组元素是否是 Data URL 格式的字符串
			if strValue, ok := item.(string); ok {
				if filename, replaced := processDataURL(strValue, outputDir); replaced {
					v[i] = filename
					continue
				}
			}

			// 递归处理嵌套结构
			if err := processImages(item, outputDir); err != nil {
				return err
			}
		}
	}

	return nil
}

// processDataURL 处理 Data URL 格式的字符串 (data:image/png;base64,...)
// 同时处理 Markdown 格式: ![image](data:image/png;base64,...)
// 返回文件名和是否成功处理的标志
func processDataURL(dataURL, outputDir string) (string, bool) {
	// 首先检查是否是 Markdown 格式: ![alt](data:image/...;base64,...)
	mdRe := regexp.MustCompile(`!\[([^\]]*)\]\(data:(image/[^;]+);base64,([^)]+)\)`)
	mdMatches := mdRe.FindStringSubmatch(dataURL)

	if len(mdMatches) == 4 {
		altText := mdMatches[1]
		mimeType := mdMatches[2]
		base64Data := mdMatches[3]

		// 保存图片
		filename, err := saveBase64Image(base64Data, mimeType, outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save markdown image: %v\n", err)
			return "", false
		}

		// 返回 Markdown 格式的文件引用
		return fmt.Sprintf("![%s](%s)", altText, filename), true
	}

	// 匹配普通 Data URL 格式: data:image/...;base64,...
	re := regexp.MustCompile(`^data:(image/[^;]+);base64,(.+)$`)
	matches := re.FindStringSubmatch(dataURL)

	if len(matches) != 3 {
		return "", false
	}

	mimeType := matches[1]
	base64Data := matches[2]

	// 保存图片
	filename, err := saveBase64Image(base64Data, mimeType, outputDir)
	if err != nil {
		// 如果保存失败，返回原值
		fmt.Fprintf(os.Stderr, "Warning: failed to save Data URL image: %v\n", err)
		return "", false
	}

	return filename, true
}

// saveBase64Image 保存 base64 编码的图片到文件
func saveBase64Image(base64Data, mimeType, outputDir string) (string, error) {
	// 解码 base64
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// 根据 mime_type 确定文件扩展名
	ext := ".png"
	if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
		ext = ".jpg"
	} else if strings.Contains(mimeType, "png") {
		ext = ".png"
	} else if strings.Contains(mimeType, "gif") {
		ext = ".gif"
	} else if strings.Contains(mimeType, "webp") {
		ext = ".webp"
	}

	// 生成文件名（使用时间戳 + 毫秒 + 计数器以避免冲突）
	now := time.Now()
	timestamp := now.Format("20060102150405") // 格式: YYYYMMDDHHMMSS
	millis := now.UnixMilli() % 1000
	counter := atomic.AddUint64(&imageCounter, 1)
	filename := fmt.Sprintf("%s%03d_%d%s", timestamp, millis, counter, ext)

	// 确定输出目录
	var decodedDir string
	if outputDir != "" {
		// 使用指定的输出目录
		decodedDir = outputDir
	} else {
		// 使用默认的 decoded 目录
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		decodedDir = filepath.Join(cwd, "decoded")
	}

	// 创建目录（如果不存在）
	if err := os.MkdirAll(decodedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// 构建完整文件路径
	fullPath := filepath.Join(decodedDir, filename)

	// 保存文件
	if err := os.WriteFile(fullPath, imageData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// 返回相对路径
	return filepath.Join(filepath.Base(decodedDir), filename), nil
}

// isImageFile 检查文件是否是图片文件
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".svg":
		return true
	}
	return false
}

// getMimeType 根据文件扩展名返回 MIME 类型
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "image/png"
	}
}

// processImageFile 处理图片文件，生成两个 base64 文件
func processImageFile(filename, outputDir string) error {
	// 读取图片文件
	imageData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read image file: %w", err)
	}

	// 编码为 base64
	base64Str := base64.StdEncoding.EncodeToString(imageData)

	// 获取 MIME 类型
	mimeType := getMimeType(filename)

	// 确定输出目录
	var dir string
	if outputDir != "" {
		// 使用指定的输出目录
		dir = outputDir
		// 创建目录（如果不存在）
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	} else {
		// 使用源文件所在目录
		dir = filepath.Dir(filename)
	}

	// 获取文件名（不含扩展名）
	baseFilename := filepath.Base(filename)
	ext := filepath.Ext(baseFilename)
	nameWithoutExt := strings.TrimSuffix(baseFilename, ext)

	// 生成两个输出文件的路径
	rawB64Path := filepath.Join(dir, nameWithoutExt+".raw.b64")
	mimeB64Path := filepath.Join(dir, nameWithoutExt+".mime.b64")

	// 写入 raw.b64 文件（纯 base64 内容）
	if err := os.WriteFile(rawB64Path, []byte(base64Str), 0644); err != nil {
		return fmt.Errorf("failed to write raw.b64 file: %w", err)
	}

	// 写入 mime.b64 文件（带 MIME 类型）
	mimeB64Content := fmt.Sprintf("%s;base64,%s", mimeType, base64Str)
	if err := os.WriteFile(mimeB64Path, []byte(mimeB64Content), 0644); err != nil {
		return fmt.Errorf("failed to write mime.b64 file: %w", err)
	}

	fmt.Printf("Generated:\n")
	fmt.Printf("  %s\n", rawB64Path)
	fmt.Printf("  %s\n", mimeB64Path)

	return nil
}

// detectImageType 检测二进制数据的图片类型，返回扩展名（如果不是图片返回空字符串）
func detectImageType(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return ".png"
	}

	// JPEG: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return ".jpg"
	}

	// GIF: 47 49 46 38 (GIF8)
	if len(data) >= 4 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return ".gif"
	}

	// WebP: RIFF....WEBP
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return ".webp"
	}

	// BMP: 42 4D (BM)
	if data[0] == 0x42 && data[1] == 0x4D {
		return ".bmp"
	}

	// SVG: starts with < or <?xml
	if data[0] == '<' {
		text := string(data[:min(100, len(data))])
		if strings.Contains(text, "<svg") || strings.Contains(text, "<?xml") {
			return ".svg"
		}
	}

	return ""
}

// isImageData 检测二进制数据是否是图片（通过文件魔数）
func isImageData(data []byte) bool {
	return detectImageType(data) != ""
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// detectImageExtension 从二进制数据检测图片扩展名
func detectImageExtension(data []byte) string {
	ext := detectImageType(data)
	if ext == "" {
		return ".png" // 默认
	}
	return ext
}

// isBase64File 检查文件是否是 base64 编码文件
func isBase64File(filename string) bool {
	// 明确的 base64 文件后缀
	if strings.HasSuffix(filename, ".mime.b64") || strings.HasSuffix(filename, ".raw.b64") {
		return true
	}

	// 检查是否是通用的 .b64 文件
	if strings.HasSuffix(filename, ".b64") {
		// 读取整个文件内容
		content, err := os.ReadFile(filename)
		if err != nil {
			return false
		}

		contentStr := string(content)

		// 检查是否包含 MIME 类型头
		var base64Sample string
		if strings.Contains(contentStr, ";base64,") {
			// 去掉 MIME 头，但只取前面的一部分用于检测
			parts := strings.SplitN(contentStr, ";base64,", 2)
			if len(parts) == 2 {
				// 只取 base64 数据的前 4096 字节用于检测（解码后约 3KB，足够检测图片魔数）
				sampleSize := min(4096, len(parts[1]))
				base64Sample = parts[1][:sampleSize]
			} else {
				return false
			}
		} else {
			// 纯 base64 内容，取前 4096 字节
			sampleSize := min(4096, len(contentStr))
			base64Sample = contentStr[:sampleSize]
		}

		// 尝试 base64 解码
		decoded, err := base64.StdEncoding.DecodeString(base64Sample)
		if err != nil {
			return false
		}

		// 检测解码后的内容是否是图片
		return isImageData(decoded)
	}

	return false
}

// decodeBase64File 解码 base64 文件并保存为图片
func decodeBase64File(filename, outputDir string) error {
	// 读取文件内容
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read base64 file: %w", err)
	}

	var base64Data string
	var mimeType string
	var ext string

	// 判断文件类型并解析
	if strings.HasSuffix(filename, ".mime.b64") {
		// mime.b64 格式: image/png;base64,iVBORw0KGgo...
		contentStr := string(content)
		parts := strings.SplitN(contentStr, ";base64,", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid mime.b64 format: expected 'mime_type;base64,data'")
		}
		mimeType = parts[0]
		base64Data = parts[1]

		// 从 MIME 类型确定扩展名
		if strings.Contains(mimeType, "png") {
			ext = ".png"
		} else if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
			ext = ".jpg"
		} else if strings.Contains(mimeType, "gif") {
			ext = ".gif"
		} else if strings.Contains(mimeType, "webp") {
			ext = ".webp"
		} else if strings.Contains(mimeType, "bmp") {
			ext = ".bmp"
		} else if strings.Contains(mimeType, "svg") {
			ext = ".svg"
		} else {
			ext = ".png" // 默认
		}
	} else if strings.HasSuffix(filename, ".raw.b64") {
		// raw.b64 格式: 纯 base64 内容
		base64Data = string(content)
		ext = ".png" // 默认扩展名，稍后会从实际数据中检测
	} else {
		// 通用 .b64 格式: 尝试作为纯 base64 或带 mime 类型的格式
		contentStr := string(content)

		// 先尝试解析为 mime 格式
		if strings.Contains(contentStr, ";base64,") {
			parts := strings.SplitN(contentStr, ";base64,", 2)
			if len(parts) == 2 {
				mimeType = parts[0]
				base64Data = parts[1]

				// 从 MIME 类型确定扩展名
				if strings.Contains(mimeType, "png") {
					ext = ".png"
				} else if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
					ext = ".jpg"
				} else if strings.Contains(mimeType, "gif") {
					ext = ".gif"
				} else if strings.Contains(mimeType, "webp") {
					ext = ".webp"
				} else if strings.Contains(mimeType, "bmp") {
					ext = ".bmp"
				} else if strings.Contains(mimeType, "svg") {
					ext = ".svg"
				} else {
					ext = ".png" // 默认
				}
			} else {
				// 作为纯 base64 处理
				base64Data = contentStr
				ext = "" // 稍后从数据中检测
			}
		} else {
			// 作为纯 base64 处理
			base64Data = contentStr
			ext = "" // 稍后从数据中检测
		}
	}

	// 解码 base64
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// 如果扩展名未确定或为默认值，从实际图片数据中检测
	if ext == "" || ext == ".png" {
		detectedExt := detectImageExtension(imageData)
		if detectedExt != ".png" || ext == "" {
			ext = detectedExt
		}
	}

	// 确定输出文件名（去掉 .mime.b64、.raw.b64 或 .b64 后缀）
	var baseFilename string
	if strings.HasSuffix(filename, ".mime.b64") {
		baseFilename = filepath.Base(strings.TrimSuffix(filename, ".mime.b64")) + ext
	} else if strings.HasSuffix(filename, ".raw.b64") {
		baseFilename = filepath.Base(strings.TrimSuffix(filename, ".raw.b64")) + ext
	} else if strings.HasSuffix(filename, ".b64") {
		baseFilename = filepath.Base(strings.TrimSuffix(filename, ".b64")) + ext
	} else {
		// 不应该到达这里，但作为后备方案
		baseFilename = filepath.Base(filename) + ext
	}

	// 确定输出目录
	var outputPath string
	if outputDir != "" {
		// 使用指定的输出目录
		// 创建目录（如果不存在）
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		outputPath = filepath.Join(outputDir, baseFilename)
	} else {
		// 使用源文件所在目录
		sourceDir := filepath.Dir(filename)
		outputPath = filepath.Join(sourceDir, baseFilename)
	}

	// 检查文件是否存在
	if _, err := os.Stat(outputPath); err == nil {
		// 文件存在，询问用户是否覆盖
		fmt.Printf("File '%s' already exists. Overwrite? (y/N): ", outputPath)
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			// 用户选择不覆盖，生成新文件名
			outputPath = generateNumberedFilename(outputPath)
		}
	}

	// 保存文件
	if err := os.WriteFile(outputPath, imageData, 0644); err != nil {
		return fmt.Errorf("failed to write image file: %w", err)
	}

	fmt.Printf("Decoded image saved to: %s\n", outputPath)
	return nil
}

// generateNumberedFilename 生成带序号的文件名（.1.png, .2.png 等）
func generateNumberedFilename(filename string) string {
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	counter := 1
	for {
		newFilename := fmt.Sprintf("%s.%d%s", nameWithoutExt, counter, ext)
		if _, err := os.Stat(newFilename); os.IsNotExist(err) {
			return newFilename
		}
		counter++
	}
}
