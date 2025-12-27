# Base64 图片双向转换工具

一个 Go 命令行工具，支持图片与 base64 格式的双向转换，从网络 URL 下载图片并编码，以及从 JSON/文本数据中提取 base64 编码的图片。

## 项目结构

```
b64/
├── src/                    # 源代码目录
│   ├── main.go            # 主入口和命令行参数处理
│   ├── encode.go          # 图片编码功能
│   ├── decode.go          # Base64 解码功能
│   ├── json.go            # JSON/文本处理功能
│   ├── download.go        # 网络下载功能
│   └── utils.go           # 工具函数（文件类型检测、MIME类型等）
├── tests/                  # 测试文件目录
│   ├── test.json
│   ├── test_dataurl.json
│   └── test_combined.json
├── build.sh               # 构建脚本
├── go.mod                 # Go 模块文件
└── README.md              # 项目文档
```

## 建议用法

将编译好的 b64 文件放到 `/usr/local/bin` 可以在任意位置调用 `b64` 命令，不用管命令的路径，和正常的 `ls`、`pwd` 等命令一样

```bash
b64 http://xx.com/path.jpg
```

## 功能特性

### 0. 网络图片下载模式（URL → Base64）

- 支持从 HTTP/HTTPS URL 下载图片
- 自动检测下载内容是否为有效图片格式
- 如果下载的不是图片，则报错且不保存文件
- **保存原始图片文件到指定目录**
- 自动生成 base64 编码文件（.raw.b64 和 .mime.b64）
- 支持所有图片格式的智能检测（PNG, JPEG, GIF, WebP, BMP, SVG）
- 可通过 `-o` 参数指定输出目录

### 1. 图片编码模式（图片 → Base64）

- 将图片文件编码为 base64 格式
- 生成两个文件：
  - `xxx.raw.b64`：纯 base64 内容
  - `xxx.mime.b64`：带 MIME 类型的完整格式（如 `image/png;base64,base64string`）
- 支持的图片格式：PNG, JPEG, GIF, WebP, BMP, SVG
- 可通过 `-o` 参数指定输出目录

### 2. Base64 解码模式（Base64 → 图片）

- 自动识别并解码 `.b64`、`.raw.b64`、`.mime.b64` 文件
- 智能检测图片类型（通过文件魔数），自动使用正确的扩展名
- 支持的格式：
  - 纯 base64 内容（`.raw.b64` 或通用 `.b64`）
  - 带 MIME 类型的格式（`.mime.b64` 或通用 `.b64`）
- 可通过 `-o` 参数指定输出目录
- 文件冲突处理：
  - 自动检测同名文件
  - 询问是否覆盖
  - 不覆盖时自动生成序号文件名（`.1.png`, `.2.png` 等）

### 3. JSON/文本处理模式

- 自动识别并提取两种格式的 base64 图片：
  1. **结构化格式**：`{"mime_type": "image/png", "data": "base64string"}`
  2. **Data URL 格式**：`"data:image/png;base64,base64string"`
- 自动根据 MIME 类型确定文件扩展名
- 将原 JSON 中的 base64 数据替换为本地文件路径
- 递归处理嵌套的 JSON 结构和数组
- 生成唯一的时间戳文件名（格式：`YYYYMMDDHHMMSSmmm_counter.ext`）

## 安装与构建

### 使用构建脚本

```bash
./build.sh
```

### 手动构建

```bash
cd src
go build -o b64 --ldflags="-s -w" --trimpath
cp -f b64 /usr/local/bins
```

## 使用方法

### 网络图片下载模式（URL → Base64）

#### 基本用法

```bash
# 从 URL 下载图片并编码为 base64（在当前目录生成）
b64 http://example.com/image.jpg
b64 https://example.com/photo.png

# 指定输出目录
b64 -o ./output http://example.com/image.jpg
b64 -o /tmp http://example.com/photo.png
```

#### 示例输出

```bash
$ b64 http://example.com/photo.jpg
Downloading from URL: http://example.com/photo.jpg
Downloaded 152340 bytes, detected as .jpg
Saved original image: photo.jpg
Generated:
  photo.raw.b64
  photo.mime.b64

$ b64 -o ./downloads http://example.com/image.png
Downloading from URL: http://example.com/image.png
Downloaded 89234 bytes, detected as .png
Saved original image: downloads/image.png
Generated:
  downloads/image.raw.b64
  downloads/image.mime.b64
```

生成的文件包括：
- **原始图片**：保持原始格式和内容
- **`.raw.b64`**：纯 base64 编码
- **`.mime.b64`**：带 MIME 类型的 base64 编码

#### 错误处理

如果 URL 指向的不是有效图片，会报错：

```bash
$ b64 http://example.com/document.pdf
Downloading from URL: http://example.com/document.pdf
Error processing URL: downloaded content is not a valid image
```

### 图片编码模式（图片 → Base64）

#### 基本用法

```bash
# 在图片所在目录生成 base64 文件
./b64 image.png

# 在当前目录生成 base64 文件
./b64 -o ./ image.png

# 在指定目录生成 base64 文件
./b64 -o /path/to/output image.png
```

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

生成的文件内容：

**photo.raw.b64**（纯 base64）：

```
/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDA...
```

**photo.mime.b64**（带 MIME 类型）：

```
image/jpeg;base64,/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDA...
```

### Base64 解码模式（Base64 → 图片）

#### 基本用法

```bash
# 解码到源文件所在目录
./b64 photo.raw.b64
# 输出: photo.png

# 解码到当前目录
./b64 -o ./ photo.mime.b64
# 输出: ./photo.png

# 解码到指定目录
./b64 -o /tmp/output photo.b64
# 输出: /tmp/output/photo.png
```

#### 智能类型检测

工具会自动检测图片类型并使用正确的扩展名：

```bash
$ ./b64 image.b64
Decoded image saved to: image.png  # 自动检测为 PNG

$ ./b64 photo.raw.b64
Decoded image saved to: photo.jpg  # 自动检测为 JPEG
```

#### 文件冲突处理

当目标文件已存在时：

```bash
$ ./b64 -o /tmp photo.raw.b64
File '/tmp/photo.png' already exists. Overwrite? (y/N): n
Decoded image saved to: /tmp/photo.1.png

$ ./b64 -o /tmp photo.raw.b64
File '/tmp/photo.png' already exists. Overwrite? (y/N): n
Decoded image saved to: /tmp/photo.2.png
```

### JSON/文本处理模式

#### 从标准输入读取

```bash
# 从 curl 获取数据
curl -s https://api.example.com/endpoint | ./b64

# 从文件读取
cat response.json | ./b64

# 使用 pretty print
cat response.json | ./b64 --pretty

# 使用 go run
cat response.json | go run main.go
```

#### 保存处理后的结果

```bash
curl -s https://api.example.com/endpoint | ./b64 > processed.json
```

#### 示例

**输入（test_combined.json）**：

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

**输出**：

```json
{
  "response": {
    "data_url_image": "decoded/20251224195004631_1.png",
    "structured_image": {
      "data": "decoded/20251224195004632_2.jpg",
      "mime_type": "image/jpeg"
    }
  }
}
```

同时在 `decoded/` 目录下生成对应的图片文件。

## 命令行参数

```
Usage: b64 [OPTIONS] [FILE]

Extract base64 encoded images from text or JSON to decoded/ directory.
Or encode image files to base64 format.
Or decode base64 files back to images.

Arguments:
  FILE                  Input file to process (reads from stdin if not provided)

Options:
  -f, --format-json     Pretty print JSON output (JSON input only)
  -p, --pretty          Pretty print JSON output (JSON input only)
  -o, --output DIR      Output directory for encoded/decoded image files
  -h, --help            Show this help message
```

### 参数说明

- **-o, --output DIR**
  - **编码模式**：指定 base64 文件的输出目录
  - **解码模式**：指定图片文件的输出目录
  - 如果不指定，默认在源文件所在目录生成文件
  - 如果指定的目录不存在，会自动创建
- **-f, --format-json / -p, --pretty**
  - 仅用于 JSON 处理模式
  - 格式化输出 JSON（带缩进）

## 支持的图片格式

| 格式 | 扩展名 | 编码 | 解码 | 自动检测 |
| ---- | ------ | ---- | ---- | -------- |
| PNG  | .png   | ✓    | ✓    | ✓        |
| JPEG | .jpg   | ✓    | ✓    | ✓        |
| GIF  | .gif   | ✓    | ✓    | ✓        |
| WebP | .webp  | ✓    | ✓    | ✓        |
| BMP  | .bmp   | ✓    | ✓    | ✓        |
| SVG  | .svg   | ✓    | ✓    | ✓        |

**自动检测**：工具通过文件魔数（magic number）自动识别图片类型，无需依赖文件扩展名。

## 文件命名规则

### JSON 处理模式

格式：`YYYYMMDDHHMMSSmmm_counter.ext`

示例：`20251224195004631_1.png`

- `20251224` - 日期（2025 年 12 月 24 日）
- `195004` - 时间（19:50:04）
- `631` - 毫秒
- `_1` - 计数器（防止冲突）
- `.png` - 扩展名

### 编码/解码模式

- **编码**：`原文件名.{raw|mime}.b64`
- **解码**：`原文件名.{png|jpg|...}`（去掉 .b64 后缀）

## 完整使用示例

### 场景 1：从网络下载图片并编码

```bash
# 1. 从 URL 下载图片并编码为 base64
$ b64 -o ./backup http://example.com/myimage.jpg
Downloading from URL: http://example.com/myimage.jpg
Downloaded 125678 bytes, detected as .jpg
Saved original image: backup/myimage.jpg
Generated:
  backup/myimage.raw.b64
  backup/myimage.mime.b64

# 2. 查看原始图片
$ open backup/myimage.jpg

# 3. 也可以解码 base64 文件验证
$ b64 -o ./restored backup/myimage.raw.b64
Decoded image saved to: restored/myimage.jpg

# 4. 验证原图和解码后的图片一致
$ diff backup/myimage.jpg restored/myimage.jpg
# (无输出表示文件相同)
```

### 场景 2：图片 → Base64 → 图片（往返转换）

```bash
# 1. 编码图片为 base64
$ ./b64 -o ./backup myimage.png
Generated:
  backup/myimage.raw.b64
  backup/myimage.mime.b64

# 2. 解码回图片
$ ./b64 -o ./restored backup/myimage.raw.b64
Decoded image saved to: restored/myimage.png

# 3. 验证文件一致
$ diff myimage.png restored/myimage.png
# (无输出表示文件相同)
```

### 场景 3：处理 API 响应中的图片

```bash
# 1. 获取包含 base64 图片的 JSON
$ curl -s https://api.example.com/images | ./b64 --pretty > output.json

# 2. 查看提取的图片
$ ls decoded/
20251226140310145_1.png
20251226140310145_2.jpg

# 3. 如需将图片重新编码
$ ./b64 -o ./encoded decoded/20251226140310145_1.png
```

### 场景 4：批量处理

```bash
# 批量编码所有 PNG 图片
for img in *.png; do
  ./b64 -o ./b64_files "$img"
done

# 批量解码所有 b64 文件
for b64 in b64_files/*.b64; do
  ./b64 -o ./restored "$b64"
done
```

## 工作原理

### 图片类型检测

工具通过检测文件的前几个字节（魔数）来识别图片类型：

| 格式 | 魔数（十六进制）                          |
| ---- | ----------------------------------------- |
| PNG  | 89 50 4E 47 0D 0A 1A 0A                   |
| JPEG | FF D8 FF                                  |
| GIF  | 47 49 46 38 (GIF8)                        |
| WebP | 52 49 46 46 ... 57 45 42 50 (RIFF...WEBP) |
| BMP  | 42 4D (BM)                                |
| SVG  | 3C (< 字符)                               |

这种方法比依赖文件扩展名更可靠。

## 测试

### 使用测试文件

测试文件位于 `tests/` 目录中：

```bash
# 测试 JSON 处理
cat tests/test_dataurl.json | ./b64
cat tests/test_combined.json | ./b64 --pretty

# 测试图片编码
./b64 test.png
./b64 -o ./output test.jpg

# 测试 base64 解码
./b64 test.raw.b64
./b64 -o ./decoded test.mime.b64
./b64 generic.b64

# 测试文件冲突
./b64 test.b64  # 第一次
./b64 test.b64  # 第二次，选择 n

# 测试网络下载
./b64 https://httpbin.org/image/png
./b64 -o ./downloads https://httpbin.org/image/jpeg
```

## 常见问题

**Q: 如何知道一个 .b64 文件是什么格式的图片？**

A: 工具会自动检测。只需运行 `./b64 file.b64`，它会自动识别图片类型并使用正确的扩展名。

**Q: 解码时生成的文件总是覆盖原文件吗？**

A: 不会。工具会询问是否覆盖。选择 "n" 后会自动生成带序号的新文件名（如 image.1.png）。

**Q: 可以处理非图片的 base64 数据吗？**

A: 工具专门设计用于图片。它会检查解码后的数据是否为有效的图片格式，非图片数据会被忽略。

**Q: 编码和解码是无损的吗？**

A: 是的。Base64 编码/解码是无损的，生成的图片与原始文件完全相同。

**Q: 从 URL 下载时如何判断是否是图片？**

A: 工具会下载文件内容后，通过文件魔数（magic number）检测图片类型。只有符合标准图片格式的文件才会被保存和编码，非图片内容会被拒绝并报错。

**Q: 网络下载支持哪些协议？**

A: 目前支持 HTTP 和 HTTPS 协议。

**Q: 下载的图片会保存在哪里？**

A: 原始图片文件会根据 `-o` 参数保存到指定目录，如果不指定则保存到当前目录。同时会在同一目录生成对应的 base64 文件（.raw.b64 和 .mime.b64）。

**Q: 从网络下载时会生成哪些文件？**

A: 会生成三个文件：
- 原始图片文件（如 image.png）
- 纯 base64 编码文件（image.raw.b64）
- 带 MIME 类型的 base64 文件（image.mime.b64）

## License

MIT
