package main

import (
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type RainbowPinwheelMask struct {
	BasePattern
	Parameters RainbowPinwheelParameters `json:"parameters"`
	Label      string                    `json:"label,omitempty"`
	currentHue float64
}

func (p *RainbowPinwheelMask) GetColorAt(point Point) Color {
	rotationDegrees := calculateAngle(point, Point{CENTER_X, CENTER_Y})
	hueVal := math.Mod(p.currentHue+rotationDegrees, MAX_HUE_VALUE)
	c := colorful.Hsv(hueVal, 1.0, 1.0)
	return Color{
		R: colorPigment(c.R * 255),
		G: colorPigment(c.G * 255),
		B: colorPigment(c.B * 255),
	}
}

func (p *RainbowPinwheelMask) Update() {
	speed := p.Parameters.Speed.Value
	reversed := p.Parameters.Reversed.Value

	if reversed {
		p.currentHue = MAX_HUE_VALUE + p.currentHue - speed
	} else {
		p.currentHue = p.currentHue + speed
	}
	p.currentHue = math.Mod(p.currentHue, MAX_HUE_VALUE)
}

func (p *RainbowPinwheelMask) GetName() string {
	return "rainbowPinwheelMask"
}

func (p *RainbowPinwheelMask) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RainbowPinwheelParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for RainbowPinwheelMask")
	}
	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}

func (p *RainbowPinwheelMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RainbowPinwheelUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *RainbowPinwheelMask) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, nil)
}
