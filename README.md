# Base64 图片提取工具

一个 Go 命令行工具，用于从 JSON 数据中提取 base64 编码的图片并保存为本地文件。

## 功能特性

- 自动识别并提取两种格式的 base64 图片：
  1. **结构化格式**：`{"mime_type": "image/png", "data": "base64string"}`
  2. **Data URL 格式**：`"data:image/png;base64,base64string"`
- 自动根据 MIME 类型确定文件扩展名（支持 PNG, JPEG, GIF, WebP）
- 将原 JSON 中的 base64 数据替换为本地文件路径
- 递归处理嵌套的 JSON 结构和数组
- 生成唯一的时间戳文件名（格式：`YYYYMMDDHHMMSSmmm.ext`）

## 安装

```bash
go build -o b64
```

## 使用方法

### 从 curl 获取数据

```bash
curl -s https://api.example.com/endpoint | ./b64
```

### 从文件读取

```bash
cat response.json | ./b64
```

### 保存处理后的结果

```bash
curl -s https://api.example.com/endpoint | ./b64 > processed.json
```

### 使用 go run

```bash
cat response.json | go run main.go
```

## 示例

### 输入（test_combined.json）

```json
{
  "response": {
    "data_url_image": "data:image/png;base64,iVBORw0KGg...",
    "structured_image": {
      "mime_type": "image/jpeg",
      "data": "/9j/4AAQSkZJRg..."
    }
  }
}
```

### 输出

```json
{
  "response": {
    "data_url_image": "20251224195004631.png",
    "structured_image": {
      "data": "20251224195004632.jpg",
      "mime_type": "image/jpeg"
    }
  }
}
```

同时在 `images/` 目录下会生成：
- `images/20251224195622798.png`
- `images/20251224195622799.jpg`

如果 `images` 目录不存在，会自动创建。

## 支持的图片格式

- PNG (`.png`)
- JPEG (`.jpg`)
- GIF (`.gif`)
- WebP (`.webp`)

## 文件命名规则

文件名格式：`YYYYMMDDHHMMSSmmm.ext`

示例：`20251224195004631.png`
- `20251224` - 日期（2025年12月24日）
- `195004` - 时间（19:50:04）
- `631` - 毫秒
- `.png` - 扩展名

## 测试

```bash
# 测试 Data URL 格式
cat test_dataurl.json | ./b64

# 测试结构化格式
cat test.json | ./b64

# 测试混合格式
cat test_combined.json | ./b64
```
