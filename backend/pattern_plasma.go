package main

import (
	"errors"
	"fmt"
	"math"
	"time"
)

type PlasmaPattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters PlasmaParameters `json:"parameters"`
	Label      string           `json:"label,omitempty"`
	time       float64
	lastUpdate time.Time
}

type PlasmaParameters struct {
	Speed      FloatParameter `json:"speed"`
	Scale      FloatParameter `json:"scale"`
	Complexity FloatParameter `json:"complexity"`
	ColorShift FloatParameter `json:"colorShift"`
}

func (p *PlasmaPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(PlasmaParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Scale.Update(newParams.Scale.Value)
	p.Parameters.Complexity.Update(newParams.Complexity.Value)
	p.Parameters.ColorShift.Update(newParams.ColorShift.Value)
	return nil
}

func (p *PlasmaPattern) Update() {
	// initialize if this is the first update
	if p.lastUpdate.IsZero() {
		p.lastUpdate = time.Now()
		p.time = 0
	}

	// calculate time delta
	now := time.Now()
	deltaTime := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now

	// update time
	p.time += deltaTime * p.Parameters.Speed.Value

	// get parameters
	scale := p.Parameters.Scale.Value
	complexity := p.Parameters.Complexity.Value
	colorShift := p.Parameters.ColorShift.Value

	// calculate plasma values for each pixel
	for i, pixel := range *p.pixelMap.pixels {
		// normalize coordinates to -1 to 1 range
		x := float64(pixel.x)/400.0 - 1.0
		y := float64(pixel.y)/400.0 - 1.0

		// calculate plasma value
		value := p.plasmaFunction(x*scale, y*scale, p.time, complexity)

		// map plasma value to color
		color := p.plasmaToColor(value, colorShift)

		// apply color mask if available
		if p.GetColorMask() != nil {
			maskColor := p.GetColorMask().GetColorAt(Point{pixel.x, pixel.y})

			// blend with mask color
			color = Color{
				R: colorPigment(float64(color.R)*0.5 + float64(maskColor.R)*0.5),
				G: colorPigment(float64(color.G)*0.5 + float64(maskColor.G)*0.5),
				B: colorPigment(float64(color.B)*0.5 + float64(maskColor.B)*0.5),
				W: 0,
			}
		}

		(*p.pixelMap.pixels)[i].color = color
	}

	// update the color mask if we have one
	if p.GetColorMask() != nil {
		p.GetColorMask().Update()
	}
}

func (p *PlasmaPattern) plasmaFunction(x, y, time, complexity float64) float64 {
	// create a plasma effect using sine waves
	v1 := math.Sin(x*complexity + time)
	v2 := math.Sin(y*complexity + time)
	v3 := math.Sin((x+y)*complexity + time)
	v4 := math.Sin(math.Sqrt(x*x+y*y)*complexity + time)

	// combine the waves
	return (v1 + v2 + v3 + v4) / 4.0
}

func (p *PlasmaPattern) plasmaToColor(value, colorShift float64) Color {
	// map plasma value (-1 to 1) to hue (0 to 360)
	hue := (value+1.0)*180.0 + colorShift
	for hue >= 360 {
		hue -= 360
	}

	// use high saturation and value
	saturation := 1.0
	value = 1.0

	// convert HSV to RGB
	r, g, b := HSVtoRGB(hue, saturation, value)

	return Color{
		R: colorPigment(r * 255),
		G: colorPigment(g * 255),
		B: colorPigment(b * 255),
		W: 0,
	}
}

func (p *PlasmaPattern) GetName() string {
	return "plasma"
}

type PlasmaUpdateRequest struct {
	Parameters PlasmaParameters `json:"parameters"`
}

func (r *PlasmaUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *PlasmaPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &PlasmaUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *PlasmaPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
