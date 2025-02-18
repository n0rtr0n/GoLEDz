package main

import (
	"errors"
	"fmt"
	"math"
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
}

func (c *Color) toString() []byte {
	return []byte{
		byte(c.R),
		byte(c.G),
		byte(c.B),
	}
}

type Gradient struct {
	StartColor Color
	EndColor   Color
}

func brightnessAdjustedColorPigment(color colorPigment, brightness float64) colorPigment {
	return colorPigment(math.Round(float64(color) * float64(brightness) / 100))
}

func brightnessAdjustedColor(color Color, brightness float64) Color {
	return Color{
		brightnessAdjustedColorPigment(color.R, brightness),
		brightnessAdjustedColorPigment(color.G, brightness),
		brightnessAdjustedColorPigment(color.B, brightness),
	}
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
	p.Value = Color{
		R: colorPigment(rand.Intn(int(MAX_PIGMENT_VALUE) + 1)),
		G: colorPigment(rand.Intn(int(MAX_PIGMENT_VALUE) + 1)),
		B: colorPigment(rand.Intn(int(MAX_PIGMENT_VALUE) + 1)),
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
	if p.Min == nil || p.Max <= *p.Min {
		return
	}
	p.Value = *p.Min + rand.Float64()*(p.Max-*p.Min)
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
	if p.Min == nil || p.Max <= *p.Min {
		return
	}
	p.Value = *p.Min + rand.Intn(p.Max-*p.Min+1)
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
	p.Value = rand.Float32() < 0.5
}
