package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// isURL 检查字符串是否是一个有效的 HTTP/HTTPS URL
func isURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// downloadFile 下载文件并返回本地临时文件路径和数据
func downloadFile(urlStr string) (string, []byte, error) {
	// 创建 HTTP 请求
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查是否是图片
	if !isImageData(data) {
		return "", nil, fmt.Errorf("downloaded content is not a valid image")
	}

	// 检测图片类型并确定扩展名
	ext := detectImageExtension(data)

	// 创建临时文件
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("b64_download_%d%s", os.Getpid(), ext))

	// 保存到临时文件
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return "", nil, fmt.Errorf("failed to write temporary file: %w", err)
	}

	return tmpFile, data, nil
}

// processURLInput 处理 URL 输入，下载文件并转换为 base64
func processURLInput(urlStr string, outputDir string) error {
	fmt.Fprintf(os.Stderr, "Downloading from URL: %s\n", urlStr)

	// 下载文件
	tmpFile, data, err := downloadFile(urlStr)
	if err != nil {
		return err
	}

	// 确保在函数返回前删除临时文件
	defer func() {
		os.Remove(tmpFile)
	}()

	// 检测文件类型
	if !isImageData(data) {
		return fmt.Errorf("downloaded file is not a valid image format")
	}

	fmt.Fprintf(os.Stderr, "Downloaded %d bytes, detected as %s\n", len(data), detectImageExtension(data))

	// 确定输出文件名
	parsedURL, _ := url.Parse(urlStr)
	baseFilename := filepath.Base(parsedURL.Path)

	// 如果 URL 路径没有文件名或没有扩展名，使用检测到的扩展名
	if baseFilename == "" || baseFilename == "/" || !isImageFile(baseFilename) {
		ext := detectImageExtension(data)
		baseFilename = fmt.Sprintf("downloaded_image%s", ext)
	} else {
		// 确保扩展名正确（基于实际内容）
		detectedExt := detectImageExtension(data)
		currentExt := strings.ToLower(filepath.Ext(baseFilename))
		if currentExt != detectedExt {
			// 替换为正确的扩展名
			baseFilename = strings.TrimSuffix(baseFilename, currentExt) + detectedExt
		}
	}

	// 确定输出目录
	var dir string
	if outputDir != "" {
		dir = outputDir
		// 创建目录（如果不存在）
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	} else {
		// 使用当前目录
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// 保存原始图片文件
	imagePath := filepath.Join(dir, baseFilename)
	if err := os.WriteFile(imagePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save original image: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Saved original image: %s\n", imagePath)

	// 使用保存的图片文件进行编码
	return processImageFile(imagePath, outputDir)
}
