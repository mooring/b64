package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
