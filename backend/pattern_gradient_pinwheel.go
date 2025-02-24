package main

// import (
// 	"errors"
// 	"fmt"
// 	"math"

// 	"github.com/lucasb-eyer/go-colorful"
// )

// type GradientPinwheelPattern struct {
// 	pixelMap          *PixelMap
// 	currentSaturation float64
// 	currentHue        float64
// 	Parameters        GradientPinwheelParameters `json:"parameters"`
// 	Label             string                     `json:"label,omitempty"`
// }

// func (p *GradientPinwheelPattern) UpdateParameters(parameters AdjustableParameters) error {
// 	newParams, ok := parameters.(GradientPinwheelParameters)
// 	if !ok {
// 		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
// 		return errors.New(err)
// 	}

// 	p.Parameters.Speed.Update(newParams.Speed.Value)
// 	p.Parameters.Divisions.Update(newParams.Divisions.Value)
// 	p.Parameters.Reversed.Update(newParams.Reversed.Value)
// 	p.Parameters.Hue.Update(newParams.Hue.Value)
// 	p.Parameters.Rainbow.Update((newParams.Rainbow.Value))
// 	return nil
// }

// //{"parameters": {"speed": {"value": 1.0},"divisions":{"value": 4},"reversed": {"value": false},"hue": {"value": 120.0}}}

// type GradientPinwheelParameters struct {
// 	Speed     FloatParameter   `json:"speed"`
// 	Divisions IntParameter     `json:"divisions"`
// 	Reversed  BooleanParameter `json:"reversed"`
// 	Hue       FloatParameter   `json:"hue"`
// 	Rainbow   BooleanParameter `json:"rainbow"`
// }

// func (p *GradientPinwheelPattern) Update() {
// 	speed := p.Parameters.Speed.Value
// 	divisions := p.Parameters.Divisions.Value
// 	reversed := p.Parameters.Reversed.Value
// 	hue := p.Parameters.Hue.Value
// 	rainbow := p.Parameters.Rainbow.Value

// 	for i, pixel := range *p.pixelMap.pixels {

// 		// this will go from 0 - 360
// 		rotationDegrees := calculateAngle(Point{pixel.x, pixel.y}, Point{CENTER_X, CENTER_Y})

// 		// we want to express this as a fraction of MAX_DEGREES, so this will be 0.0 - 1.0
// 		fractionDegrees := rotationDegrees / MAX_DEGREES * float64(divisions)

// 		// anything over 1.0 will loop back around to 0.0
// 		saturation := math.Mod(p.currentSaturation+fractionDegrees, MAX_SATURATION)

// 		if rainbow {
// 			// fmt.Println("setting current hue to")
// 			// fmt.Println(p.currentHue)
// 			hue = p.currentHue
// 		}

// 		c := colorful.Hsv(hue, saturation, 1.0)

// 		color := Color{
// 			R: colorPigment(c.R * 255),
// 			G: colorPigment(c.G * 255),
// 			B: colorPigment(c.B * 255),
// 		}
// 		(*p.pixelMap.pixels)[i].color = color
// 	}

// 	if rainbow {
// 		normalizedSpeed := 15 * speed
// 		if reversed {
// 			hue = MAX_HUE_VALUE + p.currentHue - normalizedSpeed
// 		} else {
// 			hue = p.currentHue + normalizedSpeed
// 		}
// 		p.currentHue = math.Mod(hue, MAX_HUE_VALUE)
// 	}

// 	var sat float64
// 	if reversed {
// 		// ensures that this value will not dip below 0
// 		sat = MAX_SATURATION + p.currentSaturation - speed
// 	} else {
// 		sat = p.currentSaturation + speed
// 	}

// 	p.currentSaturation = math.Mod(sat, MAX_SATURATION)
// }

// func (p *GradientPinwheelPattern) GetName() string {
// 	return "gradientPinwheel"
// }

// type GradientPinwheelUpdateRequest struct {
// 	Parameters GradientPinwheelParameters `json:"parameters"`
// }

// func (r *GradientPinwheelUpdateRequest) GetParameters() AdjustableParameters {
// 	return r.Parameters
// }

// func (p *GradientPinwheelPattern) GetPatternUpdateRequest() PatternUpdateRequest {
// 	return &GradientPinwheelUpdateRequest{
// 		Parameters: GradientPinwheelParameters{},
// 	}
// }

// func (p *GradientPinwheelPattern) TransitionFrom(source Pattern, progress float64) {
// 	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
// }
