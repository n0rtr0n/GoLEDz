package main

import (
	"errors"
	"fmt"
)

type colorPigment uint8

const MIN_PIGMENT_VALUE colorPigment = 0
const MAX_PIGMENT_VALUE colorPigment = 255

// internal parameters are set at the time the pattern is registered
// each adjustable parameter implements the update method, which
// provides validation at the time the new value is set
type AdjustableParameter interface {
	Get() interface{}
	Update(value interface{}) error
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

type ColorParameter struct {
	Value Color
}

func (p *ColorParameter) Get() interface{} {
	return p.Value
}

func (p *ColorParameter) Update(value interface{}) error {
	newValue, ok := value.(Color)
	if !ok {
		return errors.New("Invalid type for ColorParameter")
	}

	if newValue.R < MIN_PIGMENT_VALUE || newValue.R > MAX_PIGMENT_VALUE {
		return errors.New("Red color pigment provided to ColorParameter is invalid.")
	}
	if newValue.G < MIN_PIGMENT_VALUE || newValue.G > MAX_PIGMENT_VALUE {
		return errors.New("Green color pigment provided to ColorParameter is invalid.")
	}
	if newValue.B < MIN_PIGMENT_VALUE || newValue.B > MAX_PIGMENT_VALUE {
		return errors.New("Blue color pigment provided to ColorParameter is invalid.")
	}
	p.Value = newValue
	return nil
}

type FloatParameter struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

func (p *FloatParameter) Get() interface{} {
	return p.Value
}

func (p *FloatParameter) Update(value interface{}) error {
	newValue, ok := value.(float64)
	if !ok {
		return errors.New("Invalid type for FloatParameter")
	}

	if newValue < p.Min || newValue > p.Max {
		err := fmt.Sprintf(
			"Value %f provided to FloatParameter outside of range %f to %f",
			newValue,
			p.Min,
			p.Max,
		)
		return errors.New(err)
	}
	p.Value = newValue
	return nil
}

type BooleanParameter struct {
	Value bool `json:"value"`
}

func (p *BooleanParameter) Update(value interface{}) error {
	newValue, ok := value.(bool)
	if !ok {
		return errors.New("Value provided to BooleanParameter is not boolean")
	}
	p.Value = newValue
	return nil
}

func (p *BooleanParameter) Get() interface{} {
	return p.Value
}
