#!/usr/bin/env bash
# xmips 工具脚本
#   ./tool.sh build   编译 src/ 并重建系统函数库（.scp -> dist/*.co）
#   ./tool.sh clean   清除 dist/ 及子目录下所有 .co 中间产物
set -e

# 切换到脚本所在目录（项目根目录），保证相对路径可靠
cd "$(dirname "$0")"

cmd="${1:-build}"
case "$cmd" in
  build)
    echo "==> 1) 编译 src/ -> dist/xmips ..."
    (cd src && go build -o ../dist/xmips)

    echo "==> 2) 重建系统函数库 (./xmips update) ..."
    (cd dist && ./xmips update)

    echo "完成。运行："
    echo "  cd dist && ./xmips"
    echo "  直接执行单一程序: ./dist/xmips <程序名.cupa>"
    ;;
  clean)
    echo "==> 清除 dist/ 及子目录下所有 .co 文件 ..."
    find dist -name "*.co" -type f -delete
    echo "完成。清掉后再直接运行，依赖系统函数（INT 10 冒泡排序）的程序会失效；"
    echo "请先用 './tool.sh build' 重建系统函数库。"
    ;;
  *)
    echo "用法: ./tool.sh [build|clean]"
    exit 1
    ;;
esac