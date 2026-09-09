#!/usr/bin/env bash
# xmips 工具脚本（部署到系统）
#   ./tool.sh build   编译 xmips → 安装到 /usr/local/bin/xmips；dist/* 同步到 ~/.xmips/；重建系统函数库
#   ./tool.sh clean   卸载：rm -rf ~/.xmips；rm /usr/local/bin/xmips
set -e

# 切换到脚本所在目录（项目根目录），保证相对路径可靠
cd "$(dirname "$0")"

# 数据根目录：sudo 下 $HOME 会被换成 /root，这里改为落到真实调用用户的家目录，
# 保证服务用户（如 yaki）能通过默认 $HOME 找到 ~/.xmips
REAL_USER="${SUDO_USER:-$USER}"
REAL_HOME="$(getent passwd "$REAL_USER" | cut -d: -f6)"
REAL_HOME="${REAL_HOME:-$HOME}"
HOME_DIR="$REAL_HOME/.xmips"
BIN="/usr/local/bin/xmips"

cmd="${1:-build}"
case "$cmd" in
  build)
    echo "==> 1) 编译 src/xmips ..."
    TMPBIN="$(mktemp -t xmips.XXXXXX)"
    (cd src && go build -o "$TMPBIN" .)

    echo "==> 2) 安装到 $BIN ..."
    # install 一条命令完成复制+设置权限；无写权限时整体走 sudo
    if ! install -m 0755 "$TMPBIN" "$BIN" 2>/dev/null; then
      echo "    无写权限，尝试 sudo ..." && sudo install -m 0755 "$TMPBIN" "$BIN"
    fi
    rm -f "$TMPBIN" dist/xmips   # 不再把二进制留在 dist（顺带清掉旧的 dev 残留）

    echo "==> 3) 同步运行目录：dist/* -> $HOME_DIR/ ..."
    mkdir -p "$HOME_DIR"
    cp -rf dist/. "$HOME_DIR/"
    # sudo 安装时数据属主改回真实用户（如 yaki），保证服务能以该用户读写
    [[ -n "$SUDO_USER" ]] && chown -R "$REAL_USER":"$REAL_USER" "$HOME_DIR"
    chmod -R u+rwX "$HOME_DIR"

    echo "==> 4) 重建系统函数库（$HOME_DIR/sysfun/*.co）..."
    if [[ -n "$SUDO_USER" ]]; then
      # sudo 下以真实用户运行，使其按默认 $HOME 定位到 $HOME_DIR
      sudo -u "$REAL_USER" env HOME="$REAL_HOME" "$BIN" update
    else
      "$BIN" update
    fi

    echo "完成。运行：$BIN"
    echo "  数据目录：$HOME_DIR（config.ini / userfile / sysfun / run.list）"
    ;;
  clean)
    echo "==> 删除 $HOME_DIR 与 $BIN ..."
    rm -rf "$HOME_DIR"
    rm -f "$BIN"
    echo "完成。已卸载。"
    ;;
  *)
    echo "用法: ./tool.sh [build|clean]"
    exit 1
    ;;
esac