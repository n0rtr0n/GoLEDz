package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type RainbowDiagonalPattern struct {
	BasePattern
	pixelMap   *PixelMap
	currentHue float64
	Parameters RainbowDiagonalParameters `json:"parameters"`
	Label      string                    `json:"label,omitempty"`
}

func (p *RainbowDiagonalPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RainbowDiagonalParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Size.Update(newParams.Size.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}

type RainbowDiagonalParameters struct {
	Speed    FloatParameter   `json:"speed"`
	Size     FloatParameter   `json:"size"`
	Reversed BooleanParameter `json:"reversed"`
}

func (p *RainbowDiagonalPattern) Update() {
	speed := p.Parameters.Speed.Value
	size := p.Parameters.Size.Value
	reversed := p.Parameters.Reversed.Value

	for i, pixel := range *p.pixelMap.pixels {
		position := float64(pixel.x+pixel.y) * size
		hueVal := math.Mod(p.currentHue+position, MAX_HUE_VALUE)
		c := colorful.Hsv(hueVal, 1.0, 1.0)
		color := Color{
			R: colorPigment(c.R * 255),
			G: colorPigment(c.G * 255),
			B: colorPigment(c.B * 255),
		}
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

type RainbowDiagonalUpdateRequest struct {
	Parameters RainbowDiagonalParameters `json:"parameters"`
}

func (r *RainbowDiagonalUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *RainbowDiagonalPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RainbowDiagonalUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *RainbowDiagonalPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
