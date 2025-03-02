package main

import (
	"fmt"
)

type MaskOnlyPattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters MaskOnlyParameters `json:"parameters"`
	Label      string             `json:"label,omitempty"`
}

type MaskOnlyParameters struct {
	// No specific parameters needed for this pattern
}

func (p *MaskOnlyPattern) Update() {
	// Apply the color mask to all pixels
	if p.GetColorMask() != nil {
		for i, pixel := range *p.pixelMap.pixels {
			point := Point{pixel.x, pixel.y}
			(*p.pixelMap.pixels)[i].color = p.GetColorMask().GetColorAt(point)
		}
	} else {
		// Default to white if no mask is set
		for i := range *p.pixelMap.pixels {
			(*p.pixelMap.pixels)[i].color = Color{255, 255, 255, 0}
		}
	}
}

func (p *MaskOnlyPattern) GetName() string {
	return "maskOnly"
}

func (p *MaskOnlyPattern) UpdateParameters(parameters AdjustableParameters) error {
	_, ok := parameters.(MaskOnlyParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for MaskOnlyPattern")
	}
	// No parameters to update
	return nil
}

type MaskOnlyUpdateRequest struct {
	Parameters MaskOnlyParameters `json:"parameters"`
}

func (r *MaskOnlyUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *MaskOnlyPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &MaskOnlyUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *MaskOnlyPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
