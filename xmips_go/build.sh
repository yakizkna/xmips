#!/usr/bin/env bash
# 编译 Xmips（Go 复刻版）并将可执行文件拷贝到 Xmips 运行目录
set -e

# 切换到脚本所在目录（xmips_go），保证相对路径可靠
cd "$(dirname "$0")"

echo "==> 编译 src/ ..."
(cd src && go build -o ../xmips_go)

echo "==> 拷贝 xmips_go 到 Xmips/ ..."
cp xmips_go Xmips/xmips_go

echo "完成。运行："
echo "  cd Xmips && ./xmips_go"