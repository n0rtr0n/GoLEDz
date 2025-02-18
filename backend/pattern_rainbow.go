package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type RainbowPattern struct {
	pixelMap   *PixelMap
	currentHue float64
	Parameters RainbowParameters `json:"parameters"`
	Label      string            `json:"label,omitempty"`
}

func (p *RainbowPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RainbowParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Brightness.Update((newParams.Brightness.Value))
	return nil
}

type RainbowParameters struct {
	Speed      FloatParameter `json:"speed"`
	Brightness FloatParameter `json:"brightness"`
}

func (p *RainbowPattern) Update() {
	speed := p.Parameters.Speed.Value
	brightness := p.Parameters.Brightness.Value

	for i, pixel := range *p.pixelMap.pixels {
		hueVal := math.Mod(p.currentHue+float64(pixel.channelPosition), MAX_HUE_VALUE)
		c := colorful.Hsv(hueVal, 1.0, 1.0)
		color := Color{
			R: colorPigment(c.R * 255),
			G: colorPigment(c.G * 255),
			B: colorPigment(c.B * 255),
		}
		color = brightnessAdjustedColor(color, brightness)
		(*p.pixelMap.pixels)[i].color = color
	}
	p.currentHue = math.Mod(p.currentHue+speed, MAX_HUE_VALUE)
}

func (p *RainbowPattern) GetName() string {
	return "rainbow"
}

type RainbowUpdateRequest struct {
	Parameters RainbowParameters `json:"parameters"`
}

func (r *RainbowUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *RainbowPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RainbowUpdateRequest{
		Parameters: RainbowParameters{},
	}
}

func (p *RainbowPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
