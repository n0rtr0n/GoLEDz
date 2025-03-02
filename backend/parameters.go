package main

import (
	"errors"
	"fmt"
	"math/rand"
)

type colorPigment uint8

const MIN_PIGMENT_VALUE colorPigment = 0
const MAX_PIGMENT_VALUE colorPigment = 255

const MAX_BRIGHTNESS_VALUE float64 = 50.0

// internal parameters are set at the time the pattern is registered
// each adjustable parameter implements the update method, which
// provides validation at the time the new value is set
type Parameter interface {
	Get() interface{}
	Update(value interface{}) error
	Randomize()
}

type AdjustableParameters interface{}

type ParametersUpdateRequest struct {
	Parameters AdjustableParameters `json:"parameters"`
}

type Color struct {
	R colorPigment `json:"r"`
	G colorPigment `json:"g"`
	B colorPigment `json:"b"`
	W colorPigment `json:"w,omitempty"`
}

type Gradient struct {
	StartColor Color
	EndColor   Color
}

type ColorParameter struct {
	Value Color  `json:"value"`
	Type  string `json:"type,omitempty"`
}

func (p *ColorParameter) Get() interface{} {
	return p.Value
}

func (p *ColorParameter) Update(value interface{}) error {
	newValue, ok := value.(Color)
	if !ok {
		return errors.New("invalid type for ColorParameter")
	}

	if newValue.R < MIN_PIGMENT_VALUE || newValue.R > MAX_PIGMENT_VALUE {
		return errors.New("red color pigment provided to ColorParameter is invalid")
	}
	if newValue.G < MIN_PIGMENT_VALUE || newValue.G > MAX_PIGMENT_VALUE {
		return errors.New("green color pigment provided to ColorParameter is invalid")
	}
	if newValue.B < MIN_PIGMENT_VALUE || newValue.B > MAX_PIGMENT_VALUE {
		return errors.New("blue color pigment provided to ColorParameter is invalid")
	}
	p.Value = newValue
	return nil
}

func (p *ColorParameter) Randomize() {
	// Generate a random hue (0-360), high saturation (0.7-1.0), and high value (0.7-1.0)
	h := rand.Float64() * 360
	s := 0.7 + rand.Float64()*0.3
	v := 0.7 + rand.Float64()*0.3

	// Convert HSV to RGB
	r, g, b := HSVtoRGB(h, s, v)

	// Set the color
	p.Value = Color{
		R: colorPigment(r * 255),
		G: colorPigment(g * 255),
		B: colorPigment(b * 255),
	}
}

type FloatParameter struct {
	Min   *float64 `json:"min,omitempty"`
	Max   float64  `json:"max,omitempty"`
	Value float64  `json:"value"`
	Type  string   `json:"type,omitempty"`
}

func (p *FloatParameter) Get() interface{} {
	return p.Value
}

func (p *FloatParameter) Update(value interface{}) error {
	newValue, ok := value.(float64)
	if !ok {
		return errors.New("invalid type for FloatParameter")
	}

	if newValue < *p.Min || newValue > p.Max {
		err := fmt.Sprintf(
			"Value %f provided to FloatParameter outside of range %f to %f",
			newValue,
			*p.Min,
			p.Max,
		)
		return errors.New(err)
	}
	p.Value = newValue
	return nil
}

func (p *FloatParameter) Randomize() {
	if p.Min == nil {
		return // Can't randomize without min value
	}
	min := *p.Min
	rangeVal := p.Max - min
	p.Value = min + rand.Float64()*rangeVal
}

type IntParameter struct {
	Min   *int   `json:"min,omitempty"`
	Max   int    `json:"max,omitempty"`
	Value int    `json:"value"`
	Type  string `json:"type,omitempty"`
}

func (p *IntParameter) Get() interface{} {
	return p.Value
}

func (p *IntParameter) Update(value interface{}) error {
	newValue, ok := value.(int)
	if !ok {
		return errors.New("invalid type for IntParameter")
	}

	if newValue < *p.Min || newValue > p.Max {
		err := fmt.Sprintf(
			"Value %d provided to FloatParameter outside of range %d to %d",
			newValue,
			*p.Min,
			p.Max,
		)
		return errors.New(err)
	}
	p.Value = newValue
	return nil
}

func (p *IntParameter) Randomize() {
	if p.Min == nil {
		return // Can't randomize without min value
	}
	min := *p.Min
	rangeVal := p.Max - min
	p.Value = min + int(rand.Float64()*float64(rangeVal))
}

type BooleanParameter struct {
	Value bool   `json:"value"`
	Type  string `json:"type,omitempty"`
}

func (p *BooleanParameter) Get() interface{} {
	return p.Value
}

func (p *BooleanParameter) Update(value interface{}) error {
	newValue, ok := value.(bool)
	if !ok {
		return errors.New("value provided to BooleanParameter is not boolean")
	}
	p.Value = newValue
	return nil
}

func (p *BooleanParameter) Randomize() {
	p.Value = rand.Float64() > 0.5
}
