package main

import (
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

// controller-specific limitation when not running in expanded mode
const MAX_PIXEL_LENGTH = 340

type Pattern interface {
	Update()
}

type SolidColorPattern struct {
	pixelMap *PixelMap
	color    Color
}

func (p *SolidColorPattern) Update() {
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = p.color
	}
}

type SolidColorFadePattern struct {
	pixelMap   *PixelMap
	currentHue float64
	speed      float64
}

func (p *SolidColorFadePattern) Update() {
	c := colorful.Hsv(p.currentHue, 1.0, 1.0)
	color := Color{r: uint8(c.R * 255), g: uint8(c.G * 255), b: uint8(c.B * 255)}
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = color
	}
	p.currentHue = math.Mod(p.currentHue+p.speed, 360)
}

type ChaserPattern struct {
	pixelMap        *PixelMap
	color           Color
	size            uint16
	spacing         uint16
	speed           float64
	currentPosition float64
}

// TODO: add direction
// TODO: add brightness taper on either end
func (p *ChaserPattern) Update() {
	for i, pixel := range *p.pixelMap.pixels {
		if (pixel.channelPosition+uint16(p.currentPosition))%(p.size+p.spacing) < p.size {
			(*p.pixelMap.pixels)[i].color = p.color
		} else {
			(*p.pixelMap.pixels)[i].color = Color{0, 0, 0}
		}
	}
	p.currentPosition += p.speed
	p.currentPosition = math.Mod(p.currentPosition, MAX_PIXEL_LENGTH)
}

type RainbowPattern struct {
	pixelMap   *PixelMap
	currentHue float64
	speed      float64
}

func (p *RainbowPattern) Update() {
	for i, pixel := range *p.pixelMap.pixels {
		hueVal := math.Mod(p.currentHue+float64(pixel.channelPosition), 360)
		c := colorful.Hsv(hueVal, 1.0, 1.0)
		color := Color{r: uint8(c.R * 255), g: uint8(c.G * 255), b: uint8(c.B * 255)}

		(*p.pixelMap.pixels)[i].color = color
	}
	p.currentHue = math.Mod(p.currentHue+p.speed, 360)
}

type RainbowDiagonalPattern struct {
	pixelMap   *PixelMap
	currentHue float64
	speed      float64
}

// TODO: add size, direction, and orientation
func (p *RainbowDiagonalPattern) Update() {
	for i, pixel := range *p.pixelMap.pixels {
		hueVal := math.Mod(p.currentHue+float64(pixel.x+pixel.y), 360)
		c := colorful.Hsv(hueVal, 1.0, 1.0)
		color := Color{r: uint8(c.R * 255), g: uint8(c.G * 255), b: uint8(c.B * 255)}

		(*p.pixelMap.pixels)[i].color = color
	}
	p.currentHue = math.Mod(p.currentHue+p.speed, 360)
}
