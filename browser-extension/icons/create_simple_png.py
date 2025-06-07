#!/usr/bin/env python3
"""
使用纯Python创建简单的PNG图标文件
"""

import struct
import zlib

def create_simple_png(width, height, filename):
    """创建一个简单的蓝色PNG图标"""
    
    # PNG文件头
    png_signature = b'\x89PNG\r\n\x1a\n'
    
    # IHDR chunk (图像头)
    ihdr_data = struct.pack('>IIBBBBB', width, height, 8, 2, 0, 0, 0)
    ihdr_crc = zlib.crc32(b'IHDR' + ihdr_data) & 0xffffffff
    ihdr_chunk = struct.pack('>I', len(ihdr_data)) + b'IHDR' + ihdr_data + struct.pack('>I', ihdr_crc)
    
    # 创建图像数据 (简单的蓝色背景)
    image_data = bytearray()
    for y in range(height):
        image_data.append(0)  # 过滤器类型
        for x in range(width):
            # 蓝色像素 (RGB)
            image_data.extend([0x00, 0x7c, 0xba])  # #007cba
    
    # 压缩图像数据
    compressed_data = zlib.compress(bytes(image_data))
    
    # IDAT chunk (图像数据)
    idat_crc = zlib.crc32(b'IDAT' + compressed_data) & 0xffffffff
    idat_chunk = struct.pack('>I', len(compressed_data)) + b'IDAT' + compressed_data + struct.pack('>I', idat_crc)
    
    # IEND chunk (结束)
    iend_crc = zlib.crc32(b'IEND') & 0xffffffff
    iend_chunk = struct.pack('>I', 0) + b'IEND' + struct.pack('>I', iend_crc)
    
    # 写入PNG文件
    with open(filename, 'wb') as f:
        f.write(png_signature)
        f.write(ihdr_chunk)
        f.write(idat_chunk)
        f.write(iend_chunk)
    
    print(f"创建了 {filename} ({width}x{height})")

def main():
    """主函数"""
    sizes = [16, 32, 48, 128]
    
    print("正在创建简单的PNG图标...")
    
    for size in sizes:
        filename = f"icon{size}.png"
        try:
            create_simple_png(size, size, filename)
        except Exception as e:
            print(f"创建 {filename} 时出错: {e}")
    
    print("\n完成! 所有图标文件已创建。")
    print("这些是简单的蓝色方块图标，您可以稍后替换为更精美的设计。")

if __name__ == "__main__":
    main()
