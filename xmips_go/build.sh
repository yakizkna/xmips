#!/usr/bin/env bash
# 编译 Xmips（Go 复刻版）并将可执行文件拷贝到 Xmips 运行目录
set -e

# 切换到脚本所在目录（xmips_go 项目目录），保证相对路径可靠
cd "$(dirname "$0")"

# 注意：macOS 文件系统大小写不敏感，顶层文件名 "xmips" 会与运行目录 "Xmips" 冲突，
# 因此把二进制直接构建到运行目录内的 Xmips/xmips（不同层级，无冲突）。
echo "==> 编译 src/ -> Xmips/xmips ..."
(cd src && go build -o ../Xmips/xmips)

echo "完成。运行："
echo "  cd Xmips && ./xmips"