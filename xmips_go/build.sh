#!/usr/bin/env bash
# 编译 Xmips（Go 复刻版）并部署到运行目录，清空 .co 后重建系统函数库
set -e

# 切换到脚本所在目录（xmips_go 项目目录），保证相对路径可靠
cd "$(dirname "$0")"

# 注意：macOS 文件系统大小写不敏感，顶层文件名 "xmips" 会与运行目录 "Xmips" 冲突，
# 因此把二进制直接构建到运行目录内的 Xmips/xmips（不同层级，无冲突）。
echo "==> 1) 编译 src/ -> Xmips/xmips ..."
(cd src && go build -o ../Xmips/xmips)

echo "==> 2) 清空 Xmips 运行目录下的 .co 文件 ..."
(cd Xmips && rm -f ./*.co)

echo "==> 3) 重建系统函数库 (./xmips update) ..."
(cd Xmips && ./xmips update)

echo "完成。运行："
echo "  cd Xmips && ./xmips"
echo "  直接执行单一程序: ./Xmips/xmips <程序名.cupa>"