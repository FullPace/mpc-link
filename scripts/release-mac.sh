#!/bin/sh
# Builds "MPC Link.app", signs it with the Developer ID (hardened runtime), notarizes + staples it and writes
# dist/MPC-Link-<version>-macOS.zip. Signing happens in a temp folder on the system disk: on exFAT (/Volumes/Daten)
# macOS keeps xattrs as ._ files, which end up in the zip and break the signature for Apple's notary service.
#   SIGN_ID         codesign identity   (default: the "Developer ID Application" one in the keychain)
#   NOTARY_PROFILE  notarytool keychain profile (default: AC_PASSWORD)
#   SKIP_NOTARIZE=1 sign only
set -e
cd "$(dirname "$0")/.."
./scripts/make-mac-app.sh >/dev/null
VERSION=$(sed -n 's/^const version = "\(.*\)"/\1/p' cmd/mpc-link/app.go)
SIGN_ID=${SIGN_ID:-$(security find-identity -v -p codesigning | sed -n 's/.*"\(Developer ID Application: .*\)"/\1/p' | head -1)}
NOTARY_PROFILE=${NOTARY_PROFILE:-AC_PASSWORD}
[ -n "$SIGN_ID" ] || { echo "no Developer ID Application identity in the keychain" >&2; exit 1; }

STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT
APP="$STAGE/MPC Link.app"
ditto --norsrc --noextattr "build/MPC Link.app" "$APP"
codesign --force --options runtime --timestamp -s "$SIGN_ID" "$APP"
codesign --verify --strict --verbose=2 "$APP"

ZIP="$STAGE/MPC-Link-$VERSION-macOS.zip"
if [ "$SKIP_NOTARIZE" != 1 ]; then
  ditto -c -k --keepParent "$APP" "$ZIP"
  xcrun notarytool submit "$ZIP" --keychain-profile "$NOTARY_PROFILE" --wait --output-format json > "$STAGE/notary.json"
  if ! grep -q '"status" *: *"Accepted"' "$STAGE/notary.json"; then
    cat "$STAGE/notary.json" >&2
    ID=$(sed -n 's/.*"id" *: *"\([^"]*\)".*/\1/p' "$STAGE/notary.json")
    [ -n "$ID" ] && xcrun notarytool log "$ID" --keychain-profile "$NOTARY_PROFILE" >&2
    exit 1
  fi
  xcrun stapler staple "$APP"
  spctl -a -vv "$APP"
  rm -f "$ZIP"
fi
ditto -c -k --keepParent "$APP" "$ZIP"
mkdir -p dist && cp "$ZIP" dist/
echo "dist/MPC-Link-$VERSION-macOS.zip"
