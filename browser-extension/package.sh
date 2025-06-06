#!/bin/bash

# Email Server 浏览器插件打包脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${2}${1}${NC}"
}

print_message "Email Server 浏览器插件打包工具" $BLUE
echo "=================================="

# 检查当前目录
if [ ! -f "manifest.json" ]; then
    print_message "错误: 请在 browser-extension 目录中运行此脚本" $RED
    exit 1
fi

# 创建输出目录
OUTPUT_DIR="dist"
PACKAGE_NAME="email-server-extension"
VERSION=$(grep '"version"' manifest.json | sed 's/.*"version": "\([^"]*\)".*/\1/')

print_message "检测到插件版本: $VERSION" $GREEN

# 清理旧的构建文件
if [ -d "$OUTPUT_DIR" ]; then
    print_message "清理旧的构建文件..." $YELLOW
    rm -rf "$OUTPUT_DIR"
fi

mkdir -p "$OUTPUT_DIR"

# 复制必要文件
print_message "复制插件文件..." $YELLOW

# 核心文件
cp manifest.json "$OUTPUT_DIR/"
cp background.js "$OUTPUT_DIR/"
cp content.js "$OUTPUT_DIR/"
cp popup.html "$OUTPUT_DIR/"
cp popup.js "$OUTPUT_DIR/"
cp options.html "$OUTPUT_DIR/"
cp options.js "$OUTPUT_DIR/"

# 创建图标目录（如果不存在）
mkdir -p "$OUTPUT_DIR/icons"

# 复制图标文件（如果存在）
if [ -d "icons" ]; then
    cp -r icons/* "$OUTPUT_DIR/icons/" 2>/dev/null || true
fi

# 如果没有图标文件，创建占位符
if [ ! -f "$OUTPUT_DIR/icons/icon16.png" ]; then
    print_message "警告: 缺少图标文件，创建占位符..." $YELLOW
    
    # 创建简单的占位符图标（如果系统支持 ImageMagick）
    if command -v convert >/dev/null 2>&1; then
        convert -size 16x16 xc:blue "$OUTPUT_DIR/icons/icon16.png" 2>/dev/null || true
        convert -size 32x32 xc:blue "$OUTPUT_DIR/icons/icon32.png" 2>/dev/null || true
        convert -size 48x48 xc:blue "$OUTPUT_DIR/icons/icon48.png" 2>/dev/null || true
        convert -size 128x128 xc:blue "$OUTPUT_DIR/icons/icon128.png" 2>/dev/null || true
        print_message "已创建占位符图标文件" $GREEN
    else
        print_message "请手动添加图标文件到 icons/ 目录" $YELLOW
    fi
fi

# 创建压缩包
print_message "创建压缩包..." $YELLOW

cd "$OUTPUT_DIR"
ZIP_NAME="${PACKAGE_NAME}-v${VERSION}.zip"
zip -r "../$ZIP_NAME" . -x "*.DS_Store" "*.git*" "*.md"
cd ..

# 验证压缩包
if [ -f "$ZIP_NAME" ]; then
    FILE_SIZE=$(du -h "$ZIP_NAME" | cut -f1)
    print_message "✅ 打包完成!" $GREEN
    print_message "文件名: $ZIP_NAME" $GREEN
    print_message "文件大小: $FILE_SIZE" $GREEN
    echo ""
    print_message "安装说明:" $BLUE
    echo "1. 打开 Chrome 浏览器"
    echo "2. 访问 chrome://extensions/"
    echo "3. 开启'开发者模式'"
    echo "4. 拖拽 $ZIP_NAME 到页面中"
    echo "或者解压后选择 $OUTPUT_DIR 文件夹"
else
    print_message "❌ 打包失败!" $RED
    exit 1
fi

# 清理临时文件
print_message "清理临时文件..." $YELLOW
rm -rf "$OUTPUT_DIR"

print_message "🎉 所有操作完成!" $GREEN
