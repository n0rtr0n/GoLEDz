package main

import (
	"errors"
	"fmt"
	"math"
)

type ChaserPattern struct {
	BasePattern
	pixelMap        *PixelMap
	currentPosition float64
	currentHue      float64
	Parameters      ChaserParameters `json:"parameters"`
	Label           string           `json:"label,omitempty"`
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
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}

type ChaserParameters struct {
	Speed    FloatParameter   `json:"speed"`
	Size     IntParameter     `json:"size"`
	Spacing  IntParameter     `json:"spacing"`
	Reversed BooleanParameter `json:"reversed"`
}

func (p *ChaserPattern) Update() {
	speed := p.Parameters.Speed.Value
	size := p.Parameters.Size.Value
	spacing := p.Parameters.Spacing.Value
	reversed := p.Parameters.Reversed.Value

	width := uint16(size + spacing)

	for i, pixel := range *p.pixelMap.pixels {
		point := Point{pixel.x, pixel.y}
		chaserPos := pixel.channelPosition + uint16(p.currentPosition)

		if width > 0 && (chaserPos%width < uint16(size)) {
			if p.GetColorMask() != nil {
				(*p.pixelMap.pixels)[i].color = p.GetColorMask().GetColorAt(point)
			} else {
				// Default white if no color mask is set
				(*p.pixelMap.pixels)[i].color = Color{255, 255, 255, 0}
			}
		} else {
			(*p.pixelMap.pixels)[i].color = Color{0, 0, 0, 0}
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
		Parameters: p.Parameters,
	}
}

func (p *ChaserPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
