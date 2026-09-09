#!/usr/bin/env bash
# 编译 xmips（Go 复刻版）并部署到运行目录，清空 .co 后重建系统函数库
set -e

# 切换到脚本所在目录（项目根目录），保证相对路径可靠
cd "$(dirname "$0")"

echo "==> 1) 编译 src/ -> dist/xmips ..."
(cd src && go build -o ../dist/xmips)

echo "==> 2) 清空 dist 运行目录下的 .co 文件 ..."
(cd dist && rm -f ./*.co)

echo "==> 3) 重建系统函数库 (./xmips update) ..."
(cd dist && ./xmips update)

echo "完成。运行："
echo "  cd dist && ./xmips"
echo "  直接执行单一程序: ./dist/xmips <程序名.cupa>"