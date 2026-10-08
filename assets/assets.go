// Package assets holds the MPC Link logo images (rendered from SVG/black.svg by render.py) and the padbangers.com banner.
package assets

import _ "embed"

// TrayMac is the logo as a macOS template image: the black body opaque, screen and pads see-through, 44x44 px (22 pt).
//
//go:embed tray-mac.png
var TrayMac []byte

// TrayWin is the logo as drawn (white on black), 64x64 px.
//
//go:embed tray-win.png
var TrayWin []byte

// Banner is padbangers.com's header logo (white lettering, red head), 1200x180, for the settings page header.
//
//go:embed padbangers-banner.png
var Banner []byte
