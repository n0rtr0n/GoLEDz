package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type ChaserPattern struct {
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
	p.Parameters.Color.Update(newParams.Color.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	p.Parameters.Rainbow.Update(newParams.Rainbow.Value)
	return nil
}

type ChaserParameters struct {
	Speed    FloatParameter   `json:"speed"`
	Size     IntParameter     `json:"size"`
	Spacing  IntParameter     `json:"spacing"`
	Color    ColorParameter   `json:"color"`
	Reversed BooleanParameter `json:"reversed"`
	Rainbow  BooleanParameter `json:"rainbow"`
}

func (p *ChaserPattern) Update() {
	speed := p.Parameters.Speed.Value
	size := p.Parameters.Size.Value
	spacing := p.Parameters.Spacing.Value
	color := p.Parameters.Color.Value
	reversed := p.Parameters.Reversed.Value
	rainbow := p.Parameters.Rainbow.Value

	width := uint16(size + spacing)

	for i, pixel := range *p.pixelMap.pixels {
		if rainbow {
			hueVal := math.Mod(p.currentHue+float64(pixel.channelPosition), MAX_HUE_VALUE)
			c := colorful.Hsv(hueVal, 1.0, 1.0)
			color = Color{
				R: colorPigment(c.R * 255),
				G: colorPigment(c.G * 255),
				B: colorPigment(c.B * 255),
			}
		}

		chaserPos := pixel.channelPosition + uint16(p.currentPosition)
		if width > 0 && (chaserPos%width < uint16(size)) {
			(*p.pixelMap.pixels)[i].color = color
		} else {
			(*p.pixelMap.pixels)[i].color = Color{0, 0, 0}
		}
	}

	p.currentHue = math.Mod(p.currentHue+speed, MAX_HUE_VALUE)

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

func (p *ChaserPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
