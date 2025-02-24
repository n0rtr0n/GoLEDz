package main

import (
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type RainbowColorMask struct {
	BasePattern
	Parameters RainbowParameters `json:"parameters"`
	Label      string            `json:"label,omitempty"`
	currentHue float64
}

func (p *RainbowColorMask) GetColorAt(point Point) Color {
	hueVal := math.Mod(p.currentHue+float64(point.X+point.Y), MAX_HUE_VALUE)
	c := colorful.Hsv(hueVal, 1.0, 1.0)
	return Color{
		R: colorPigment(c.R * 255),
		G: colorPigment(c.G * 255),
		B: colorPigment(c.B * 255),
	}
}

func (p *RainbowColorMask) Update() {
	speed := p.Parameters.Speed.Value
	p.currentHue = math.Mod(p.currentHue+speed, MAX_HUE_VALUE)
}

func (p *RainbowColorMask) GetName() string {
	return "rainbowColorMask"
}

func (p *RainbowColorMask) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RainbowParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for RainbowColorMask")
	}
	p.Parameters.Speed.Update(newParams.Speed.Value)
	return nil
}

func (p *RainbowColorMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RainbowUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *RainbowColorMask) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, nil)
}
