package main

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// effectively pi
const MAX_CURSOR float64 = 180.0

type PulsePattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters PulseParameters `json:"parameters"`
	Label      string          `json:"label,omitempty"`
}

func (p *PulsePattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(PulseParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.MinBrightness.Update(newParams.MinBrightness.Value)
	p.Parameters.MaxBrightness.Update(newParams.MaxBrightness.Value)
	return nil
}

type PulseParameters struct {
	Speed         FloatParameter `json:"speed"`
	MinBrightness FloatParameter `json:"minBrightness"`
	MaxBrightness FloatParameter `json:"maxBrightness"`
}

func (p *PulsePattern) Update() {
	// calculate the current brightness based on time
	t := time.Now().UnixNano() / int64(time.Millisecond)
	speed := p.Parameters.Speed.Value

	// convert frequency to milliseconds for the sine wave
	period := 1000.0 / speed

	// calculate normalized position in the sine wave (0.0 to 1.0)
	position := float64(t%int64(period)) / period

	// use sine wave to create pulsing effect (0.0 to 1.0)
	brightness := (math.Sin(position*2*math.Pi) + 1) / 2

	// apply min/max brightness range
	minBrightness := p.Parameters.MinBrightness.Value / 100.0
	maxBrightness := p.Parameters.MaxBrightness.Value / 100.0
	brightness = minBrightness + brightness*(maxBrightness-minBrightness)

	for i, pixel := range *p.pixelMap.pixels {
		var pixelColor Color

		// if we don't have a color mask, do nothing
		if p.GetColorMask() == nil {
			return
		}
		point := Point{pixel.x, pixel.y}
		baseColor := p.GetColorMask().GetColorAt(point)

		// apply brightness to the color from the mask
		pixelColor = Color{
			R: colorPigment(float64(baseColor.R) * brightness),
			G: colorPigment(float64(baseColor.G) * brightness),
			B: colorPigment(float64(baseColor.B) * brightness),
		}

		(*p.pixelMap.pixels)[i].color = pixelColor
	}
}

func (p *PulsePattern) GetName() string {
	return "pulse"
}

type PulseUpdateRequest struct {
	Parameters PulseParameters `json:"parameters"`
}

func (r *PulseUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *PulsePattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &PulseUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *PulsePattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
