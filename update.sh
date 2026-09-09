#!/usr/bin/env bash
# xmips 部署脚本（同时承担初始化与更新），供 ops_console 一键更新调用。
#
# 职责（幂等，重复执行安全）：
#   1) git pull 拉最新代码；
#   2) yaki 全量配置：确保 ~/.xmips/config.ini 为 fileEnable=1 / sockEnable=1；
#   3) tool.sh build：编译安装 xmips + 同步 dist/* 到 yaki 的 ~/.xmips/ + 重建系统函数库；
#   4) 受限账号 xmipsuser 环境：确保用户存在，建 ~/.xmips，同步 yaki 数据，
#      并用 config.ini.user 替换其 config.ini（fileEnable/sockEnable 强制为 0）；
#   5) 重建 xmipsuser 系统函数库；
#   6) sudoers：允许 yaki 免密以 xmipsuser 仅执行 /usr/local/bin/xmips；
#   7) yakisite unit：注入 Environment=XMIPS_PUBLIC_USER=xmipsuser。
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="/usr/local/bin/xmips"
XUSER="xmipsuser"
YAKI_HM="/home/yaki/.xmips"
X_HM="/home/${XUSER}/.xmips"

echo "[1/7] git pull"
cd "$REPO_DIR"
HEAD_BEFORE="$(git rev-parse HEAD)"
git pull --ff-only origin master
HEAD_AFTER="$(git rev-parse HEAD)"

# 本次发生了版本更新，而 bash 读取的仍是旧脚本 → exec 重载新版本再继续，
# 避免用旧逻辑执行后续步骤（如旧的系统函数库重建 / env 注入写法）。
if [ "$HEAD_BEFORE" != "$HEAD_AFTER" ]; then
    echo "  [reload] 版本更新 $HEAD_BEFORE → $HEAD_AFTER，重新加载脚本"
    exec bash "$0" "$@"
fi

echo "[2/7] yaki 全量配置（file/socket 开）"
mkdir -p "$YAKI_HM"
YCFG="$YAKI_HM/config.ini"
[ -f "$YCFG" ] || { cp dist/config.ini "$YCFG"; }
sed -i -E 's/^fileEnable=.*/fileEnable=1/; s/^sockEnable=.*/sockEnable=1/' "$YCFG"
grep -q '^fileEnable=' "$YCFG" || echo 'fileEnable=1' >> "$YCFG"
grep -q '^sockEnable=' "$YCFG" || echo 'sockEnable=1' >> "$YCFG"
echo "  yaki  $YCFG:"; grep -E '^(fileEnable|sockEnable)=' "$YCFG"

echo "[3/7] tool.sh build（编译安装 + 同步 yaki ~/.xmips + 重建系统函数库）"
./tool.sh build

echo "[4/7] 受限账号 $XUSER 数据目录（file/socket 关）"
id "$XUSER" >/dev/null 2>&1 || sudo useradd -m -s /bin/bash "$XUSER"
sudo mkdir -p "$X_HM/userfile" "$X_HM/sysfun"
# 清理历史遗留的嵌套路径（旧脚本误用 env HOME 时留下的 ~/.xmips/.xmips）
sudo rm -rf "$X_HM/.xmips"
sudo cp -rf "$YAKI_HM/userfile/." "$X_HM/userfile/"
sudo cp -rf "$YAKI_HM/sysfun/." "$X_HM/sysfun/"
if [ -f "$YAKI_HM/config.ini.user" ]; then
    echo "  [config] 用 config.ini.user 替换 $XUSER 的 config.ini"
    sudo cp -f "$YAKI_HM/config.ini.user" "$X_HM/config.ini"
else
    echo "  [config] 未见 config.ini.user，沿用 yaki 的 config.ini"
    sudo cp -f "$YCFG" "$X_HM/config.ini"
    sudo sed -i -E 's/^fileEnable=.*/fileEnable=0/; s/^sockEnable=.*/sockEnable=0/' "$X_HM/config.ini"
fi
sudo chown -R "$XUSER":"$XUSER" "$X_HM"
sudo chmod -R u+rwX "$X_HM"
echo "  $X_HM/config.ini:"; sudo grep -E '^(fileEnable|sockEnable)=' "$X_HM/config.ini"

echo "[5/7] 重建受限账号系统函数库"
# 不要覆盖 HOME：sudo -u 会使用 xmipsuser 真实家目录(/home/xmipsuser)，xmips 据此定位 ~/.xmips
sudo -u "$XUSER" "$BIN" update

echo "[6/7] sudoers：yaki 免密以 $XUSER 执行 $BIN"
SUDO_FILE="/etc/sudoers.d/xmips-plus"
sudo bash -c "cat > '$SUDO_FILE' <<'SUDEOF'
Defaults:yaki !requiretty
yaki ALL=($XUSER) NOPASSWD: $BIN
SUDEOF"
sudo chmod 0440 "$SUDO_FILE"
sudo visudo -c -f "$SUDO_FILE" || { echo 'visudo 校验失败'; sudo rm -f "$SUDO_FILE"; exit 1; }
echo "  $SUDO_FILE OK"

echo "[7/7] yakisite unit 注入 XMIPS_PUBLIC_USER"
UNIT_D="/etc/systemd/system/yakisite.service.d"
sudo mkdir -p "$UNIT_D"
UNIT_F="$(sudo ls "$UNIT_D"/*.conf 2>/dev/null | head -1)"
[ -n "$UNIT_F" ] || UNIT_F="$UNIT_D/override.conf"
# 确保 [Service] 段头存在
if ! sudo grep -qx '\[Service\]' "$UNIT_F" 2>/dev/null; then
    sudo bash -c "{ echo '[Service]'; :; } > '$UNIT_F.tmp' && cat '$UNIT_F' >> '$UNIT_F.tmp' && mv '$UNIT_F.tmp' '$UNIT_F'"
fi
# 注入环境变量（幂等）
sudo bash -c "grep -qx 'Environment=XMIPS_PUBLIC_USER=$XUSER' '$UNIT_F' || echo 'Environment=XMIPS_PUBLIC_USER=$XUSER' >> '$UNIT_F'"
echo "  $UNIT_F:"; sudo grep -E '^\[' "$UNIT_F"; sudo grep '^Environment=' "$UNIT_F" || true

echo "[8/7] 验证"
sudo -u "$XUSER" "$BIN" 2>&1 | head -n 3 || true
echo "完成。xmips=$BIN  xmipsuser=$X_HM"