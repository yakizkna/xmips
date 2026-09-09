#!/usr/bin/env bash
# xmips 工具脚本
#   ./tool.sh build   编译 src/、重建系统函数库（.scp -> dist/sysfun/*.co）、并安装到 ~/.xmips/dist
#   ./tool.sh clean   清除 dist 下除 sysfun/ 外的 .co 中间产物（系统函数 .co 保留以加速）
set -e

# 切换到脚本所在目录（项目根目录），保证相对路径可靠
cd "$(dirname "$0")"

INSTALL_DIR="$HOME/.xmips/dist"

cmd="${1:-build}"
case "$cmd" in
  build)
    echo "==> 1) 编译 src/ -> dist/xmips ..."
    (cd src && go build -o ../dist/xmips)

    echo "==> 2) 重建系统函数库（写入 dist/sysfun/*.co）..."
    (cd dist && ./xmips update)

    echo "==> 3) 安装到 $INSTALL_DIR ..."
    mkdir -p "$INSTALL_DIR"
    cp -rf dist/. "$INSTALL_DIR/"

    echo "完成。安装目录：$INSTALL_DIR"
    echo "  运行：$INSTALL_DIR/xmips"
    echo "  系统函数 .co 位于 $INSTALL_DIR/sysfun/，运行时直接读取以加速"
    ;;
  clean)
    echo "==> 清除 dist 下除 sysfun/ 外的 .co 文件 ..."
    find dist -name "*.co" -not -path "dist/sysfun/*" -delete
    echo "完成。系统函数库 .co（dist/sysfun/）已保留以加速运行。"
    ;;
  *)
    echo "用法: ./tool.sh [build|clean]"
    exit 1
    ;;
esac