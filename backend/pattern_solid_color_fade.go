package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type SolidColorFadePattern struct {
	pixelMap   *PixelMap
	Parameters SolidColorFadeParameters `json:"parameters"`
	currentHue float64
}

func (p *SolidColorFadePattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(SolidColorFadeParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Color.Update(newParams.Color.Value)
	return nil
}

type SolidColorFadeParameters struct {
	Speed FloatParameter `json:"speed"`
	Color ColorParameter `json:"color"`
}

func (p *SolidColorFadePattern) Update() {
	speed := p.Parameters.Speed.Value

	c := colorful.Hsv(p.currentHue, 1.0, 1.0)
	color := Color{R: colorPigment(c.R * 255), G: colorPigment(c.G * 255), B: colorPigment(c.B * 255)}
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = color
	}
	p.currentHue = math.Mod(p.currentHue+speed, MAX_HUE_VALUE)
}

func (p *SolidColorFadePattern) GetName() string {
	return "solidColorFade"
}

type SolidColorFadeUpdateRequest struct {
	Parameters SolidColorFadeParameters `json:"parameters"`
}

func (r *SolidColorFadeUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *SolidColorFadePattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &SolidColorFadeUpdateRequest{
		Parameters: SolidColorFadeParameters{},
	}
}
