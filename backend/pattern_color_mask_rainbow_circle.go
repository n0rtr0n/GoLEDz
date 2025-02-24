package main

import (
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type RainbowCircleMask struct {
	BasePattern
	Parameters RainbowCircleParameters `json:"parameters"`
	Label      string                  `json:"label,omitempty"`
	currentHue float64
}

func (p *RainbowCircleMask) GetColorAt(point Point) Color {
	distance := math.Sqrt(math.Pow(float64(CENTER_X-point.X), 2) + math.Pow(float64(CENTER_Y-point.Y), 2))
	hueVal := math.Mod(p.currentHue+distance, MAX_HUE_VALUE)
	c := colorful.Hsv(hueVal, 1.0, 1.0)
	return Color{
		R: colorPigment(c.R * 255),
		G: colorPigment(c.G * 255),
		B: colorPigment(c.B * 255),
	}
}

func (p *RainbowCircleMask) Update() {
	speed := p.Parameters.Speed.Value
	reversed := p.Parameters.Reversed.Value

	if reversed {
		p.currentHue = MAX_HUE_VALUE + p.currentHue - speed
	} else {
		p.currentHue = p.currentHue + speed
	}
	p.currentHue = math.Mod(p.currentHue, MAX_HUE_VALUE)
}

func (p *RainbowCircleMask) GetName() string {
	return "rainbowCircleMask"
}

func (p *RainbowCircleMask) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RainbowCircleParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for RainbowCircleMask")
	}
	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}

func (p *RainbowCircleMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RainbowCircleUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *RainbowCircleMask) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, nil)
}
