package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type RainbowCirclePattern struct {
	BasePattern
	pixelMap   *PixelMap
	currentHue float64
	Parameters RainbowCircleParameters `json:"parameters"`
	Label      string                  `json:"label,omitempty"`
}

func (p *RainbowCirclePattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RainbowCircleParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Size.Update(newParams.Size.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}

type RainbowCircleParameters struct {
	Speed    FloatParameter   `json:"speed"`
	Size     FloatParameter   `json:"size"`
	Reversed BooleanParameter `json:"reversed"`
}

func (p *RainbowCirclePattern) Update() {
	speed := p.Parameters.Speed.Value
	// size := p.Parameters.Size.Value
	reversed := p.Parameters.Reversed.Value

	for i, pixel := range *p.pixelMap.pixels {
		distance := math.Sqrt(math.Pow(float64(CENTER_X-pixel.x), 2) + math.Pow(float64(CENTER_Y-pixel.y), 2))

		hueVal := math.Mod(p.currentHue+distance, MAX_HUE_VALUE)
		c := colorful.Hsv(hueVal, 1.0, 1.0)
		color := Color{
			R: colorPigment(c.R * 255),
			G: colorPigment(c.G * 255),
			B: colorPigment(c.B * 255),
		}
		(*p.pixelMap.pixels)[i].color = color
	}

	var hue float64
	if reversed {
		// ensures that this value will not dip below 0
		hue = MAX_HUE_VALUE + p.currentHue - speed
	} else {
		hue = p.currentHue + speed
	}

	p.currentHue = math.Mod(hue, MAX_HUE_VALUE)
}

func (p *RainbowCirclePattern) GetName() string {
	return "rainbowCircle"
}

type RainbowCircleUpdateRequest struct {
	Parameters RainbowCircleParameters `json:"parameters"`
}

func (r *RainbowCircleUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *RainbowCirclePattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RainbowCircleUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *RainbowCirclePattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
