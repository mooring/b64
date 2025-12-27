package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
