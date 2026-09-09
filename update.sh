#!/usr/bin/env bash
# xmips 部署脚本：拉最新代码 -> 编译安装 -> 重建系统函数库 -> 同步给受限账号 xmipsuser
# 用法: ./update.sh
# 说明:
#   - 供 ops_console 一键更新调用，等同 deploy_xmips.exp 的 xmips 部分。
#   - 编译安装交给 tool.sh build，会把 dist/* 同步到 yaki 的 ~/.xmips/。
#   - 随后把 yaki 的 ~/.xmips 同步到受限账号 xmipsuser，并用 config.ini.user
#     替换其 config.ini（fileEnable/sockEnable 强制为 0），再重建其系统函数库。
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="/usr/local/bin/xmips"
XUSER="xmipsuser"

echo "[1/5] git pull"
cd "$REPO_DIR"
git pull --ff-only origin master

echo "[2/5] tool.sh build（编译安装 + 同步 yaki ~/.xmips + 重建系统函数库）"
./tool.sh build

echo "[3/5] 同步到受限账号 $XUSER 的 ~/.xmips"
YAKI_HM="/home/yaki/.xmips"
X_HM="/home/${XUSER}/.xmips"
sudo mkdir -p "$X_HM/userfile" "$X_HM/sysfun"
sudo cp -rf "$YAKI_HM/userfile/." "$X_HM/userfile/"
sudo cp -rf "$YAKI_HM/sysfun/." "$X_HM/sysfun/"
if [ -f "$YAKI_HM/config.ini.user" ]; then
    echo "  [config] 用 config.ini.user 替换 $XUSER 的 config.ini"
    sudo cp -f "$YAKI_HM/config.ini.user" "$X_HM/config.ini"
else
    echo "  [config] 未见 config.ini.user，沿用 yaki 的 config.ini"
    sudo cp -f "$YAKI_HM/config.ini" "$X_HM/config.ini"
    sudo sed -i -E 's/^fileEnable=.*/fileEnable=0/; s/^sockEnable=.*/sockEnable=0/' "$X_HM/config.ini"
fi
sudo chown -R "$XUSER":"$XUSER" "$X_HM"
sudo chmod -R u+rwX "$X_HM"
echo "  $X_HM/config.ini:"; sudo grep -E '^(fileEnable|sockEnable)=' "$X_HM/config.ini"

echo "[4/5] 重建受限账号系统函数库"
sudo -u "$XUSER" env HOME="$X_HM" "$BIN" update

echo "[5/5] 验证"
sudo -u "$XUSER" env HOME="$X_HM" "$BIN" 2>&1 | head -n 3 || true
echo "完成。xmips=$BIN  xmipsuser=$X_HM"