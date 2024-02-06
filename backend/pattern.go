package main

import (
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

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
