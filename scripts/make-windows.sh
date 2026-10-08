#!/bin/sh
# Builds "MPC Link.exe" for Windows (amd64 and arm64) into build/, with the app icon and no console window.
set -e
cd "$(dirname "$0")/.."
GO=${GO:-go}
# Windows resources (icon + manifest) as .syso files next to main; go build links them in.
(cd cmd/mpc-link && $GO run github.com/tc-hib/go-winres@v0.3.3 simply --icon ../../assets/appicon-1024.png \
  --manifest gui --product-name "MPC Link" --file-description "MPC Link" \
  --product-version "$(sed -n 's/^const version = "\(.*\)"/\1/p' app.go)" --arch amd64,arm64)
for arch in amd64 arm64; do
  GOOS=windows GOARCH=$arch CGO_ENABLED=0 $GO build -trimpath -ldflags "-s -w -H windowsgui" \
    -o "build/MPC Link-$arch.exe" ./cmd/mpc-link
done
rm -f cmd/mpc-link/rsrc_windows_*.syso
ls -la build/*.exe
