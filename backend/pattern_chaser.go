package main

import (
	"errors"
	"fmt"
	"math"
)

type ChaserPattern struct {
	pixelMap        *PixelMap
	currentPosition float64
	Parameters      ChaserParameters `json:"parameters"`
}

func (p *ChaserPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(ChaserParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Size.Update(newParams.Size.Value)
	p.Parameters.Spacing.Update(newParams.Spacing.Value)
	p.Parameters.Color.Update(newParams.Color.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}

type ChaserParameters struct {
	Speed    FloatParameter   `json:"speed"`
	Size     IntParameter     `json:"size"`
	Spacing  IntParameter     `json:"spacing"`
	Color    ColorParameter   `json:"color"`
	Reversed BooleanParameter `json:"reversed"`
}

func (p *ChaserPattern) Update() {
	speed := p.Parameters.Speed.Value
	size := p.Parameters.Size.Value
	spacing := p.Parameters.Spacing.Value
	color := p.Parameters.Color.Value
	reversed := p.Parameters.Reversed.Value

	width := uint16(size + spacing)

	for i, pixel := range *p.pixelMap.pixels {

		chaserPos := pixel.channelPosition + uint16(p.currentPosition)
		if width > 0 && (chaserPos%width < uint16(size)) {
			(*p.pixelMap.pixels)[i].color = color
		} else {
			(*p.pixelMap.pixels)[i].color = Color{0, 0, 0}
		}
	}
	if reversed {
		// ensures that this value will not dip below 0
		p.currentPosition = MAX_PIXEL_LENGTH + p.currentPosition - speed
	} else {
		p.currentPosition = p.currentPosition + speed
	}
	p.currentPosition = math.Mod(p.currentPosition, MAX_PIXEL_LENGTH)
}

func (p *ChaserPattern) GetName() string {
	return "chaser"
}

type ChaserUpdateRequest struct {
	Parameters ChaserParameters `json:"parameters"`
}

func (r *ChaserUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *ChaserPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &ChaserUpdateRequest{
		Parameters: ChaserParameters{},
	}
}
