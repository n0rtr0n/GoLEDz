package main

import (
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

// controller-specific limitation when not running in expanded mode
const MAX_PIXEL_LENGTH = 340
const MAX_HUE_VALUE = 360

// abitrary for now; we'll calculate this later
const MAX_X_POSITION = 600

// TODO: return and handle any errors encountered in updating patterns
type Pattern interface {
	ListParameters() *AdjustableParameters
	Update()
	GetName() string
}

type SolidColorPattern struct {
	pixelMap   *PixelMap
	parameters AdjustableParameters
}

func (p *SolidColorPattern) ListParameters() *AdjustableParameters {
	return &p.parameters
}

func (p *SolidColorPattern) Update() {
	// While I'm not a fan of this explicit cast here, I'm not sure of any other elegant method
	color := p.parameters["color"].Get().(Color)
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = color
	}
}

func (p *SolidColorPattern) GetName() string {
	return "solidColor"
}

type SolidColorFadePattern struct {
	pixelMap   *PixelMap
	parameters AdjustableParameters
	currentHue float64
}

func (p *SolidColorFadePattern) ListParameters() *AdjustableParameters {
	return &p.parameters
}

func (p *SolidColorFadePattern) Update() {
	speed := p.parameters["speed"].Get().(float64)

	c := colorful.Hsv(p.currentHue, 1.0, 1.0)
	color := Color{R: colorPigment(c.R * 255), G: colorPigment(c.G * 255), B: colorPigment(c.B * 255)}
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = color
	}
	p.currentHue = math.Mod(p.currentHue+speed, MAX_HUE_VALUE)
}

func (p *SolidColorFadePattern) GetName() string {
	return "solidColorFade"
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

func (p *ChaserPattern) GetName() string {
	return "chaser"
}

type RainbowPattern struct {
	pixelMap   *PixelMap
	currentHue float64
	parameters AdjustableParameters
}

func (p *RainbowPattern) ListParameters() *AdjustableParameters {
	return &p.parameters
}

func (p *RainbowPattern) Update() {
	speed := p.parameters["speed"].Get().(float64)
	for i, pixel := range *p.pixelMap.pixels {
		hueVal := math.Mod(p.currentHue+float64(pixel.channelPosition), MAX_HUE_VALUE)
		c := colorful.Hsv(hueVal, 1.0, 1.0)
		color := Color{R: colorPigment(c.R * 255), G: colorPigment(c.G * 255), B: colorPigment(c.B * 255)}
		(*p.pixelMap.pixels)[i].color = color
	}
	p.currentHue = math.Mod(p.currentHue+speed, MAX_HUE_VALUE)
}

func (p *RainbowPattern) GetName() string {
	return "rainbow"
}

type RainbowDiagonalPattern struct {
	pixelMap   *PixelMap
	currentHue float64
	parameters AdjustableParameters
}

func (p *RainbowDiagonalPattern) ListParameters() *AdjustableParameters {
	return &p.parameters
}

// TODO: add orientation
func (p *RainbowDiagonalPattern) Update() {
	speed := p.parameters["speed"].Get().(float64)
	size := p.parameters["size"].Get().(float64)
	reversed := p.parameters["reversed"].Get().(bool)

	for i, pixel := range *p.pixelMap.pixels {
		position := float64(pixel.x+pixel.y) * size
		hueVal := math.Mod(p.currentHue+position, MAX_HUE_VALUE)
		c := colorful.Hsv(hueVal, 1.0, 1.0)
		color := Color{R: colorPigment(c.R * 255), G: colorPigment(c.G * 255), B: colorPigment(c.B * 255)}
		(*p.pixelMap.pixels)[i].color = color
	}

	var hue float64
	if reversed {
		// ensures that this value will not dip below 0
		hue = MAX_HUE_VALUE + p.currentHue - speed
	} else {
		hue = p.currentHue + speed
	}

	p.currentHue = math.Mod(hue, MAX_HUE_VALUE)
}

func (p *RainbowDiagonalPattern) GetName() string {
	return "rainbowDiagonal"
}

type VerticalStripesPattern struct {
	pixelMap        *PixelMap
	parameters      AdjustableParameters
	currentPosition float64
}

func (p *VerticalStripesPattern) ListParameters() *AdjustableParameters {
	return &p.parameters
}

func (p *VerticalStripesPattern) Update() {
	speed := p.parameters["speed"].Get().(float64)
	size := p.parameters["size"].Get().(float64)
	color := p.parameters["color"].Get().(Color)

	min := int16(p.currentPosition - size)
	max := int16(p.currentPosition + size)
	for i, pixel := range *p.pixelMap.pixels {
		if pixel.x > min && pixel.x < max {
			(*p.pixelMap.pixels)[i].color = color
		} else {
			(*p.pixelMap.pixels)[i].color = Color{}
		}
	}

	p.currentPosition = math.Mod(p.currentPosition+speed, MAX_X_POSITION)
}

func (p *VerticalStripesPattern) GetName() string {
	return "verticalStripes"
}
