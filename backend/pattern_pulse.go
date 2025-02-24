package main

import (
	"errors"
	"fmt"
	"math"
)

// effectively pi
const MAX_CURSOR float64 = 180.0

type PulsePattern struct {
	BasePattern
	pixelMap   *PixelMap
	cursor     float64
	Parameters PulseParameters `json:"parameters"`
	Label      string          `json:"label,omitempty"`
}

func (p *PulsePattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(PulseParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Color.Update(newParams.Color.Value)
	return nil
}

type PulseParameters struct {
	Speed FloatParameter `json:"speed"`
	Color ColorParameter `json:"color"`
}

func (p *PulsePattern) Update() {
	color := p.Parameters.Color.Value
	speed := p.Parameters.Speed.Value

	p.cursor = math.Mod(p.cursor+(speed), MAX_CURSOR)
	brightness := math.Sin(degreesToRadians(p.cursor)) * 100

	for i, _ := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = brightnessAdjustedColor(color, brightness)
	}
}

func (p *PulsePattern) GetName() string {
	return "pulse"
}

type PulseUpdateRequest struct {
	Parameters PulseParameters `json:"parameters"`
}

func (r *PulseUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *PulsePattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &PulseUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *PulsePattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
