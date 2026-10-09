#!/bin/sh
# Builds "MPC Link.app" (universal: arm64 + x86_64) into build/. A menu bar app: LSUIElement, no Dock icon.
set -e
cd "$(dirname "$0")/.."
GO=${GO:-go}
VERSION=$(sed -n 's/^const version = "\(.*\)"/\1/p' cmd/mpc-link/app.go)
APP="build/MPC Link.app"
rm -rf "$APP" && mkdir -p "$APP/Contents/MacOS"
# without this, clang (cgo) stamps the build Mac's own macOS version as the minimum into the binary.
# 13.0 = what the Go runtime (Go 1.27) itself needs; the flags also go into Go's build cache key.
MINOS=13.0
export MACOSX_DEPLOYMENT_TARGET=$MINOS CGO_CFLAGS="-mmacosx-version-min=$MINOS" CGO_LDFLAGS="-mmacosx-version-min=$MINOS"
CGO_ENABLED=1 GOARCH=arm64 $GO build -trimpath -ldflags "-s -w" -o build/pbl-arm64 ./cmd/mpc-link
CGO_ENABLED=1 GOARCH=amd64 $GO build -trimpath -ldflags "-s -w" -o build/pbl-amd64 ./cmd/mpc-link
lipo -create -output "$APP/Contents/MacOS/mpc-link" build/pbl-arm64 build/pbl-amd64
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
  <key>CFBundleName</key><string>MPC Link</string>
  <key>CFBundleDisplayName</key><string>MPC Link</string>
  <key>CFBundleIdentifier</key><string>com.padbangers.mpclink</string>
  <key>CFBundleExecutable</key><string>mpc-link</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>$VERSION</string>
  <key>CFBundleVersion</key><string>$VERSION</string>
  <key>LSMinimumSystemVersion</key><string>$MINOS</string>
  <key>CFBundleIconFile</key><string>AppIcon</string>
  <key>LSUIElement</key><true/>
  <key>NSHighResolutionCapable</key><true/>
</dict></plist>
PLIST
codesign --force --deep -s - "$APP" 2>/dev/null || true   # ad-hoc signature (no Apple developer account)
echo "$APP"
