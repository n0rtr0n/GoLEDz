package main

import (
	"errors"
	"fmt"
)

type SolidColorPattern struct {
	pixelMap   *PixelMap
	Parameters SolidColorParameters `json:"parameters"`
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

type SolidColorParameters struct {
	Color ColorParameter `json:"color"`
}

func (p *SolidColorPattern) Update() {
	color := p.Parameters.Color.Value
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = color
	}
}

func (p *SolidColorPattern) GetName() string {
	return "solidColor"
}

type SolidColorUpdateRequest struct {
	Parameters SolidColorParameters `json:"parameters"`
}

func (r *SolidColorUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *SolidColorPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &SolidColorUpdateRequest{
		Parameters: SolidColorParameters{},
	}
}
