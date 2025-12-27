package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

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
