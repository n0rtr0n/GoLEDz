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
	Update(value interface{}) error
}

type AdjustastableParameterList struct {
	parameters map[string]AdjustableParameter
}

type BooleanParameter struct {
	value bool
}

func (p *BooleanParameter) Update(value bool) error {
	if value != true && value == false {
		return errors.New("Value provided to BooleanParameter is not boolean")
	}
	p.value = value
	return nil
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

func (p *FloatParameter) Update(value float64) error {

	if p.value < p.min || p.value > p.max {
		err := fmt.Sprintf(
			"Value %f provided to IntegerParameter outside of range %f to %f",
			value,
			p.min,
			p.max,
		)
		return errors.New(err)
	}
	p.value = value
	return nil
}

type Color struct {
	r colorPigment
	g colorPigment
	b colorPigment
}

func (c *Color) toString() []byte {
	return []byte{
		byte(c.r),
		byte(c.g),
		byte(c.b),
	}
}

type ColorParameter struct {
	value Color
}

func (p *ColorParameter) Update(value Color) error {
	if value.r < MIN_PIGMENT_VALUE || value.r > MAX_PIGMENT_VALUE {
		return errors.New("Red color pigment provided to ColorParameter is invalid.")
	}
	if value.g < MIN_PIGMENT_VALUE || value.g > MAX_PIGMENT_VALUE {
		return errors.New("Green color pigment provided to ColorParameter is invalid.")
	}
	if value.b < MIN_PIGMENT_VALUE || value.b > MAX_PIGMENT_VALUE {
		return errors.New("Blue color pigment provided to ColorParameter is invalid.")
	}
	p.value = value
	return nil
}
