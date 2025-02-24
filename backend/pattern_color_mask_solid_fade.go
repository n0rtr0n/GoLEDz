package main

import (
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type SolidColorFadeMask struct {
	BasePattern
	Parameters SolidColorFadeParameters `json:"parameters"`
	Label      string                   `json:"label,omitempty"`
	currentHue float64
}

func (p *SolidColorFadeMask) GetColorAt(point Point) Color {
	c := colorful.Hsv(p.currentHue, 1.0, 1.0)
	return Color{
		R: colorPigment(c.R * 255),
		G: colorPigment(c.G * 255),
		B: colorPigment(c.B * 255),
	}
}

func (p *SolidColorFadeMask) Update() {
	speed := p.Parameters.Speed.Value
	p.currentHue = math.Mod(p.currentHue+speed, MAX_HUE_VALUE)
}

func (p *SolidColorFadeMask) GetName() string {
	return "solidColorFadeMask"
}

func (p *SolidColorFadeMask) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(SolidColorFadeParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for SolidColorFadeMask")
	}
	p.Parameters.Speed.Update(newParams.Speed.Value)
	return nil
}

func (p *SolidColorFadeMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return &SolidColorFadeUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *SolidColorFadeMask) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, nil)
}
