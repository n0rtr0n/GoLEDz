package main

import (
	"errors"
	"fmt"
	"math"
)

type VerticalStripesPattern struct {
	pixelMap        *PixelMap
	currentPosition float64
	Parameters      VerticalStripesParameters `json:"parameters"`
}

func (p *VerticalStripesPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(VerticalStripesParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Size.Update(newParams.Size.Value)
	p.Parameters.Color.Update(newParams.Color.Value)
	return nil
}

type VerticalStripesParameters struct {
	Speed FloatParameter `json:"speed"`
	Size  FloatParameter `json:"size"`
	Color ColorParameter `json:"color"`
}

func (p *VerticalStripesPattern) Update() {
	speed := p.Parameters.Speed.Value
	size := p.Parameters.Size.Value
	color := p.Parameters.Color.Value

	min := int16(p.currentPosition - size)
	max := int16(p.currentPosition + size)
	for i, pixel := range *p.pixelMap.pixels {
		if pixel.x > min && pixel.x < max {
			(*p.pixelMap.pixels)[i].color = color
		} else {
			(*p.pixelMap.pixels)[i].color = Color{}
		}
	}

	p.currentPosition = math.Mod(p.currentPosition+speed, MAX_X_POSITION)
}

func (p *VerticalStripesPattern) GetName() string {
	return "verticalStripes"
}

type VerticalStripesUpdateRequest struct {
	Parameters VerticalStripesParameters `json:"parameters"`
}

func (r *VerticalStripesUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *VerticalStripesPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &VerticalStripesUpdateRequest{
		Parameters: VerticalStripesParameters{},
	}
}
