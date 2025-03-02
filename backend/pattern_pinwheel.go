package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type PinwheelPattern struct {
	BasePattern
	pixelMap          *PixelMap
	currentSaturation float64
	Parameters        PinwheelParameters `json:"parameters"`
	Label             string             `json:"label,omitempty"`
}

func (p *PinwheelPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(PinwheelParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Divisions.Update(newParams.Divisions.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}

type PinwheelParameters struct {
	Speed     FloatParameter   `json:"speed"`
	Divisions IntParameter     `json:"divisions"`
	Reversed  BooleanParameter `json:"reversed"`
}

func (p *PinwheelPattern) Update() {
	speed := p.Parameters.Speed.Value
	divisions := p.Parameters.Divisions.Value
	reversed := p.Parameters.Reversed.Value

	for i, pixel := range *p.pixelMap.pixels {
		point := Point{pixel.x, pixel.y}

		// calculate rotation degrees
		rotationDegrees := calculateAngle(point, Point{CENTER_X, CENTER_Y})

		// calculate saturation based on rotation
		fractionDegrees := rotationDegrees / MAX_DEGREES * float64(divisions)
		saturation := math.Mod(p.currentSaturation+fractionDegrees, MAX_SATURATION)

		if p.GetColorMask() != nil {
			baseColor := p.GetColorMask().GetColorAt(point)

			// convert to HSV, modify saturation, convert back
			h, _, v := RGBtoHSV(float64(baseColor.R)/255, float64(baseColor.G)/255, float64(baseColor.B)/255)
			s := saturation // apply pinwheel saturation effect
			r, g, b := HSVtoRGB(h, s, v)

			(*p.pixelMap.pixels)[i].color = Color{
				R: colorPigment(r * 255),
				G: colorPigment(g * 255),
				B: colorPigment(b * 255),
			}
		} else {
			// Default to white with saturation effect if no mask
			c := colorful.Hsv(0, saturation, 1.0) // White hue with varying saturation
			(*p.pixelMap.pixels)[i].color = Color{
				R: colorPigment(c.R * 255),
				G: colorPigment(c.G * 255),
				B: colorPigment(c.B * 255),
			}
		}
	}

	// Update saturation position
	if reversed {
		p.currentSaturation = MAX_SATURATION + p.currentSaturation - speed
	} else {
		p.currentSaturation = p.currentSaturation + speed
	}
	p.currentSaturation = math.Mod(p.currentSaturation, MAX_SATURATION)
}

func (p *PinwheelPattern) GetName() string {
	return "pinwheel"
}

type PinwheelUpdateRequest struct {
	Parameters PinwheelParameters `json:"parameters"`
}

func (r *PinwheelUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *PinwheelPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &PinwheelUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *PinwheelPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
