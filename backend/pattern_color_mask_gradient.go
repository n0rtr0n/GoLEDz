package main

import (
	"fmt"
	"math"
)

type GradientColorMask struct {
	BasePattern
	Parameters   GradientParameters `json:"parameters"`
	Label        string             `json:"label,omitempty"`
	currentAngle float64
}

func (p *GradientColorMask) GetColorAt(point Point) Color {
	color1 := p.Parameters.Color1.Value
	color2 := p.Parameters.Color2.Value
	return GetColorAtPoint(point, color1, color2, p.currentAngle)
}

func (p *GradientColorMask) Update() {
	speed := p.Parameters.Speed.Value
	reversed := p.Parameters.Reversed.Value

	if reversed {
		p.currentAngle += speed
	} else {
		p.currentAngle = (p.currentAngle - speed) + MAX_DEGREES
	}
	p.currentAngle = math.Mod(p.currentAngle, MAX_DEGREES)
}

func (p *GradientColorMask) GetName() string {
	return "gradientColorMask"
}

type GradientColorMaskUpdateRequest struct {
	Parameters GradientParameters `json:"parameters"`
}

func (r *GradientColorMaskUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *GradientColorMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return &GradientColorMaskUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *GradientColorMask) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, nil)
}

func (p *GradientColorMask) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(GradientParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for GradientColorMask")
	}
	p.Parameters.Color1.Update(newParams.Color1.Value)
	p.Parameters.Color2.Update(newParams.Color2.Value)
	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}
