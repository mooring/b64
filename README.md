# Base64 图片提取工具

一个 Go 命令行工具，用于从 JSON 数据中提取 base64 编码的图片并保存为本地文件，同时也支持将图片文件编码为 base64 格式。

## 功能特性

### JSON/文本处理模式
- 自动识别并提取两种格式的 base64 图片：
  1. **结构化格式**：`{"mime_type": "image/png", "data": "base64string"}`
  2. **Data URL 格式**：`"data:image/png;base64,base64string"`
- 自动根据 MIME 类型确定文件扩展名（支持 PNG, JPEG, GIF, WebP）
- 将原 JSON 中的 base64 数据替换为本地文件路径
- 递归处理嵌套的 JSON 结构和数组
- 生成唯一的时间戳文件名（格式：`YYYYMMDDHHMMSSmmm.ext`）

### 图片编码模式
- 将图片文件编码为 base64 格式
- 在相同目录生成两个文件：
  - `xxx.raw.b64`：纯 base64 内容
  - `xxx.mime.b64`：带 MIME 类型的完整格式（如 `image/png;base64,base64string`）
- 支持的图片格式：PNG, JPEG, GIF, WebP, BMP, SVG
- 自动从生成的文件名中去除原扩展名

## 安装

```bash
go build -o b64
```

## 使用方法

### JSON/文本处理模式

#### 从 curl 获取数据

```bash
curl -s https://api.example.com/endpoint | ./b64
```

#### 从文件读取

```bash
cat response.json | ./b64
```

#### 保存处理后的结果

```bash
curl -s https://api.example.com/endpoint | ./b64 > processed.json
```

#### 使用 go run

```bash
cat response.json | go run main.go
```

### 图片编码模式

#### 将图片文件编码为 base64

```bash
# 在图片所在目录生成 base64 文件
./b64 image.png

# 在当前目录生成 base64 文件
./b64 -o ./ image.png

# 在指定目录生成 base64 文件
./b64 -o /path/to/output image.png
```

这将生成两个文件：
- `xxx.raw.b64` - 纯 base64 内容
- `xxx.mime.b64` - 带 MIME 类型的完整格式（如 `image/png;base64,base64string`）

注意：如果原文件有扩展名（如 `.png`），生成的文件名会自动去除该扩展名。

#### 命令行参数

- `-o, --output DIR` - 指定输出目录（仅用于图片编码模式）
  - 如果不指定，默认在图片所在目录生成文件
  - 如果指定的目录不存在，会自动创建

#### 示例输出

```bash
$ ./b64 photo.jpg
Generated:
  photo.raw.b64
  photo.mime.b64

$ ./b64 -o ./output photo.jpg
Generated:
  output/photo.raw.b64
  output/photo.mime.b64
```

`photo.raw.b64` 内容：
```
/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDA...
```

`photo.mime.b64` 内容：
```
image/jpeg;base64,/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDA...
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
