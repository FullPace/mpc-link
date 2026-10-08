package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"runtime"
)

// iconPNG draws the menu bar icon: two linked rings, black on transparent (a macOS template image; Windows shows
// it as is). connected adds a filled dot.
func iconPNG(connected bool) []byte {
	const n = 44 // drawn at 2x, shown at 22 pt
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	ring := func(cx, cy, r, w float64) {
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				d := math.Abs(math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy) - r)
				if a := math.Max(0, math.Min(1, w/2+0.5-d)); a > 0 {
					old := img.RGBAAt(x, y).A
					img.SetRGBA(x, y, color.RGBA{0, 0, 0, uint8(math.Max(float64(old), a*255))})
				}
			}
		}
	}
	ring(16, 22, 9, 3.6)
	ring(28, 22, 9, 3.6)
	if connected {
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				if d := math.Hypot(float64(x)+0.5-37, float64(y)+0.5-9); d < 5.5 {
					img.SetRGBA(x, y, color.RGBA{0, 0, 0, 255})
				}
			}
		}
	}
	var b bytes.Buffer
	png.Encode(&b, img)
	return b.Bytes()
}

// icon returns the icon in the format the platform's tray wants: PNG, or on Windows an ICO wrapping the PNG.
func icon(connected bool) []byte {
	p := iconPNG(connected)
	if runtime.GOOS != "windows" {
		return p
	}
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, []uint16{0, 1, 1})                 // ICONDIR: reserved, type icon, 1 image
	b.Write([]byte{44, 44, 0, 0})                                              // width, height, colours, reserved
	binary.Write(&b, binary.LittleEndian, []uint16{1, 32})                    // planes, bit count
	binary.Write(&b, binary.LittleEndian, []uint32{uint32(len(p)), 6 + 16})   // size, offset
	b.Write(p)
	return b.Bytes()
}
