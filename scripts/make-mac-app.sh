#!/bin/sh
# Builds "Padbangers Link.app" (universal: arm64 + x86_64) into build/. A menu bar app: LSUIElement, no Dock icon.
set -e
cd "$(dirname "$0")/.."
GO=${GO:-go}
VERSION=$(sed -n 's/^const version = "\(.*\)"/\1/p' cmd/padbangers-link/app.go)
APP="build/Padbangers Link.app"
rm -rf "$APP" && mkdir -p "$APP/Contents/MacOS"
CGO_ENABLED=1 GOARCH=arm64 $GO build -trimpath -ldflags "-s -w" -o build/pbl-arm64 ./cmd/padbangers-link
CGO_ENABLED=1 GOARCH=amd64 $GO build -trimpath -ldflags "-s -w" -o build/pbl-amd64 ./cmd/padbangers-link
lipo -create -output "$APP/Contents/MacOS/padbangers-link" build/pbl-arm64 build/pbl-amd64
rm -f build/pbl-arm64 build/pbl-amd64
# app icon: the red head (assets/appicon-1024.png) at every size macOS asks for
ICONSET=build/AppIcon.iconset; rm -rf $ICONSET; mkdir -p $ICONSET "$APP/Contents/Resources"
for s in 16 32 128 256 512; do
  sips -z $s $s assets/appicon-1024.png --out $ICONSET/icon_${s}x${s}.png >/dev/null
  sips -z $((s*2)) $((s*2)) assets/appicon-1024.png --out $ICONSET/icon_${s}x${s}@2x.png >/dev/null
done
iconutil -c icns $ICONSET -o "$APP/Contents/Resources/AppIcon.icns" && rm -rf $ICONSET
cat > "$APP/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>CFBundleName</key><string>Padbangers Link</string>
  <key>CFBundleDisplayName</key><string>Padbangers Link</string>
  <key>CFBundleIdentifier</key><string>com.padbangers.link</string>
  <key>CFBundleExecutable</key><string>padbangers-link</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>$VERSION</string>
  <key>CFBundleVersion</key><string>$VERSION</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>CFBundleIconFile</key><string>AppIcon</string>
  <key>LSUIElement</key><true/>
  <key>NSHighResolutionCapable</key><true/>
</dict></plist>
PLIST
codesign --force --deep -s - "$APP" 2>/dev/null || true   # ad-hoc signature (no Apple developer account)
echo "$APP"
