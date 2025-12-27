package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

var imageCounter uint64

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

// generateTimestampFilename 生成带时间戳的唯一文件名
func generateTimestampFilename(ext string) string {
	now := time.Now()
	timestamp := now.Format("20060102150405") // 格式: YYYYMMDDHHMMSS
	millis := now.UnixMilli() % 1000
	counter := atomic.AddUint64(&imageCounter, 1)
	return fmt.Sprintf("%s%03d_%d%s", timestamp, millis, counter, ext)
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

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

	// 生成文件名
	filename := generateTimestampFilename(ext)

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
