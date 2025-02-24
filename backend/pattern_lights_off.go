package main

import (
	"errors"
	"fmt"
)

type LightsOffPattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters LightsOffParameters `json:"parameters"`
	Label      string              `json:"label,omitempty"`
}

type LightsOffParameters struct{}

func (p *LightsOffPattern) UpdateParameters(parameters AdjustableParameters) error {
	_, ok := parameters.(LightsOffParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}
	return nil
}

func (p *LightsOffPattern) Update() {
	color := Color{
		R: 0,
		G: 0,
		B: 0,
	}
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = color
	}
}

func (p *LightsOffPattern) GetName() string {
	return "lightsOff"
}

type LightsOffUpdateRequest struct {
	Parameters LightsOffParameters `json:"parameters"`
}

func (r *LightsOffUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *LightsOffPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &LightsOffUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *LightsOffPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
