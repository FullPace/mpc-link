package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"runtime"

	"github.com/FullPace/padbangers-link/assets"
)

// icon returns the tray icon in the platform's format. macOS: the black logo as a template image (the system tints
// it), faded when not connected. Windows: the red head as an ICO, grey when not connected.
func icon(connected bool) []byte {
	if runtime.GOOS == "darwin" {
		if connected {
			return assets.TrayMac
		}
		return recolor(assets.TrayMac, func(c color.NRGBA) color.NRGBA { c.A = uint8(int(c.A) * 40 / 100); return c })
	}
	p := assets.TrayWin
	if !connected {
		p = recolor(p, func(c color.NRGBA) color.NRGBA {
			g := uint8((int(c.R)*30 + int(c.G)*59 + int(c.B)*11) / 100)
			return color.NRGBA{g, g, g, c.A}
		})
	}
	if runtime.GOOS == "windows" {
		return ico(p, 64)
	}
	return p
}

// recolor applies f to every pixel of a PNG.
func recolor(src []byte, f func(color.NRGBA) color.NRGBA) []byte {
	img, err := png.Decode(bytes.NewReader(src))
	if err != nil {
		return src
	}
	b := img.Bounds()
	out := image.NewNRGBA(b)
	draw.Draw(out, b, img, b.Min, draw.Src)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			out.SetNRGBA(x, y, f(out.NRGBAAt(x, y)))
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, out)
	return buf.Bytes()
}

// ico wraps a square PNG in an ICO container (Windows' tray wants ICO; an ICO may hold PNG data).
func ico(p []byte, size int) []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, []uint16{0, 1, 1})               // reserved, type icon, 1 image
	b.Write([]byte{byte(size), byte(size), 0, 0})                           // width, height, colours, reserved
	binary.Write(&b, binary.LittleEndian, []uint16{1, 32})                  // planes, bits per pixel
	binary.Write(&b, binary.LittleEndian, []uint32{uint32(len(p)), 6 + 16}) // data size, data offset
	b.Write(p)
	return b.Bytes()
}
