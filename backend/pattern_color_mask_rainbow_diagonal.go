package main

import (
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type RainbowDiagonalMask struct {
	BasePattern
	Parameters RainbowDiagonalParameters `json:"parameters"`
	Label      string                    `json:"label,omitempty"`
	currentHue float64
}

func (p *RainbowDiagonalMask) GetColorAt(point Point) Color {
	size := p.Parameters.Size.Value
	position := float64(point.X+point.Y) * size
	hueVal := math.Mod(p.currentHue+position, MAX_HUE_VALUE)
	c := colorful.Hsv(hueVal, 1.0, 1.0)
	return Color{
		R: colorPigment(c.R * 255),
		G: colorPigment(c.G * 255),
		B: colorPigment(c.B * 255),
	}
}

func (p *RainbowDiagonalMask) Update() {
	speed := p.Parameters.Speed.Value
	reversed := p.Parameters.Reversed.Value

	if reversed {
		p.currentHue = MAX_HUE_VALUE + p.currentHue - speed
	} else {
		p.currentHue = p.currentHue + speed
	}
	p.currentHue = math.Mod(p.currentHue, MAX_HUE_VALUE)
}

func (p *RainbowDiagonalMask) GetName() string {
	return "rainbowDiagonalMask"
}

func (p *RainbowDiagonalMask) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RainbowDiagonalParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for RainbowDiagonalMask")
	}
	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Size.Update(newParams.Size.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}

func (p *RainbowDiagonalMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RainbowDiagonalUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *RainbowDiagonalMask) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, nil)
}
