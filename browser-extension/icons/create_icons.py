#!/usr/bin/env python3
"""
创建浏览器扩展所需的图标文件
"""

import os
import sys

def create_simple_icon(size, filename):
    """创建简单的SVG图标并转换为PNG"""
    
    # 创建SVG内容
    svg_content = f'''<?xml version="1.0" encoding="UTF-8"?>
<svg width="{size}" height="{size}" viewBox="0 0 {size} {size}" xmlns="http://www.w3.org/2000/svg">
  <circle cx="{size//2}" cy="{size//2}" r="{size//2-1}" fill="#007cba" stroke="#005a8b" stroke-width="1"/>
  <text x="{size//2}" y="{size//2+size//6}" text-anchor="middle" fill="white" font-family="Arial, sans-serif" font-size="{size//2}" font-weight="bold">E</text>
</svg>'''
    
    # 保存SVG文件
    svg_filename = filename.replace('.png', '.svg')
    with open(svg_filename, 'w') as f:
        f.write(svg_content)
    
    print(f"创建了 {svg_filename}")
    
    # 尝试使用不同的工具转换为PNG
    conversion_commands = [
        f"rsvg-convert -w {size} -h {size} {svg_filename} -o {filename}",
        f"inkscape --export-png={filename} --export-width={size} --export-height={size} {svg_filename}",
        f"convert {svg_filename} -resize {size}x{size} {filename}",
    ]
    
    for cmd in conversion_commands:
        if os.system(f"{cmd} 2>/dev/null") == 0:
            print(f"成功创建 {filename}")
            return True
    
    print(f"警告: 无法转换 {svg_filename} 为 PNG 格式")
    print("请手动安装 rsvg-convert, inkscape 或 imagemagick")
    return False

def main():
    """主函数"""
    sizes = [16, 32, 48, 128]
    
    print("正在创建浏览器扩展图标...")
    
    success_count = 0
    for size in sizes:
        filename = f"icon{size}.png"
        if create_simple_icon(size, filename):
            success_count += 1
    
    print(f"\n完成! 成功创建了 {success_count}/{len(sizes)} 个PNG图标")
    
    if success_count == 0:
        print("\n临时解决方案:")
        print("1. 手动下载一些简单的图标文件")
        print("2. 或者安装图像转换工具: sudo apt install librsvg2-bin")
        print("3. 然后重新运行此脚本")

if __name__ == "__main__":
    main()
