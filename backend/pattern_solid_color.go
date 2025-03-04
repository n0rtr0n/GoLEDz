package main

import (
	"errors"
	"fmt"
)

type SolidColorPattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters SolidColorParameters `json:"parameters"`
	Label      string
}

type SolidColorParameters struct {
	Color ColorParameter `json:"color"`
}

func (p *SolidColorPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(SolidColorParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}
	p.Parameters.Color.Update(newParams.Color.Value)
	return nil
}

func (p *SolidColorPattern) Update() {
	// Get the color from parameters
	color := p.Parameters.Color.Value

	// Create a new color with W explicitly set to 0
	safeColor := Color{
		R: color.R,
		G: color.G,
		B: color.B,
		W: 0, // Always force W to 0
	}

	// Apply to all pixels
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = safeColor
	}
}

func (p *SolidColorPattern) GetName() string {
	return "solidColor"
}

func (p *SolidColorPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &SolidColorUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *SolidColorPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}

type SolidColorUpdateRequest struct {
	Parameters SolidColorParameters `json:"parameters"`
}

func (r *SolidColorUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}
