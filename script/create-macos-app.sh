#!/bin/bash

# 为 macOS 创建 .app 应用包

APP_NAME="Git Manage Service"
APP_DIR="/tmp/GitManageService.app"
CONTENTS_DIR="$APP_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"

# 清理旧的构建
rm -rf "$APP_DIR"

# 创建应用包结构
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# 复制可执行文件
cp output/git-manage-service "$MACOS_DIR/"
chmod +x "$MACOS_DIR/git-manage-service"

# 创建启动脚本
cat > "$MACOS_DIR/launcher" <<'EOF'
#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"
./git-manage-service --mode=all &
SERVER_PID=$!
sleep 3
open "http://localhost:38080"
wait $SERVER_PID
EOF
chmod +x "$MACOS_DIR/launcher"

# 创建 Info.plist
cat > "$CONTENTS_DIR/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>launcher</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.yi-nology.git-manage-service</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>0.8.0</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

# 复制配置文件
cp -r conf "$RESOURCES_DIR/"

# 创建 DMG（可选）
echo "应用包已创建: $APP_DIR"
echo ""
echo "测试运行:"
echo "  open $APP_DIR"
echo ""
echo "创建 DMG:"
echo "  hdiutil create -volname '$APP_NAME' -srcfolder '$APP_DIR' -ov -format UDZO GitManageService.dmg"
