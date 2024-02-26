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

type AdjustableParameters map[string]AdjustableParameter

type BooleanParameter struct {
	value bool
}

func (p *BooleanParameter) Update(value interface{}) error {
	newValue, ok := value.(bool)
	if !ok {
		return errors.New("Value provided to BooleanParameter is not boolean")
	}
	p.value = newValue
	return nil
}

func (p *BooleanParameter) Get() interface{} {
	return p.value
}

type IntegerParameter struct {
	min   int64
	max   int64
	value int64
}

func (p *IntegerParameter) Update(value int64) error {
	if p.value < p.min || p.value > p.max {
		err := fmt.Sprintf(
			"Value %d provided to IntegerParameter outside of range %d to %d",
			value,
			p.min,
			p.max,
		)
		return errors.New(err)
	}
	p.value = value
	return nil
}

type FloatParameter struct {
	min   float64
	max   float64
	value float64
}

func (p *FloatParameter) Get() interface{} {
	return p.value
}

func (p *FloatParameter) Update(value interface{}) error {
	newValue, ok := value.(float64)
	if !ok {
		return errors.New("Invalid type for FloatParameter")
	}

	if newValue < p.min || newValue > p.max {
		err := fmt.Sprintf(
			"Value %f provided to FloatParameter outside of range %f to %f",
			newValue,
			p.min,
			p.max,
		)
		return errors.New(err)
	}
	p.value = newValue
	return nil
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
	value Color
}

func (p *ColorParameter) Get() interface{} {
	return p.value
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
	p.value = newValue
	return nil
}
