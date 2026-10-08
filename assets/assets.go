// Package assets holds the Padbangers logo images (rendered from the SVGs here by render.py).
package assets

import _ "embed"

// TrayMac is the whole logo, black on transparent, 72x44 px (36x22 pt): a macOS template image.
//
//go:embed tray-mac.png
var TrayMac []byte

// TrayWin is the red head with its headphone band, 64x64 px.
//
//go:embed tray-win.png
var TrayWin []byte
