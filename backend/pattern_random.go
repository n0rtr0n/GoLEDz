package main

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

type RandomPattern struct {
	pixelMap          *PixelMap
	Label             string           `json:"label,omitempty"`
	Parameters        RandomParameters `json:"parameters"`
	patterns          map[string]Pattern
	currentPattern    Pattern
	isActive          bool
	lastPatternSwitch time.Time
}

type RandomParameters struct {
	SwitchInterval FloatParameter `json:"switchInterval"`
}

func (p *RandomPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RandomParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}
	p.Parameters = newParams
	return nil
}

func (p *RandomPattern) Update() {
	if !p.isActive {
		return
	}

	// Check if it's time to switch patterns
	if time.Since(p.lastPatternSwitch).Seconds() >= p.Parameters.SwitchInterval.Value {
		p.switchToRandomPattern()
		p.lastPatternSwitch = time.Now()
	}

	// Update the current pattern
	if p.currentPattern != nil {
		p.currentPattern.Update()
	}
}

func (p *RandomPattern) switchToRandomPattern() {
	// Get all pattern names except our own
	var availablePatterns []Pattern
	for name, pattern := range p.patterns {
		if name != p.GetName() && name != "lightsOff" {
			availablePatterns = append(availablePatterns, pattern)
		}
	}

	// Pick a random pattern
	if len(availablePatterns) > 0 {
		newPattern := availablePatterns[rand.Intn(len(availablePatterns))]
		p.currentPattern = newPattern

		// Randomize its parameters
		p.randomizeParameters()
	}
}

func (p *RandomPattern) randomizeParameters() {
	if p.currentPattern == nil {
		return
	}

	// Get the parameters struct from the current pattern
	params := reflect.ValueOf(p.currentPattern).Elem().FieldByName("Parameters")
	if !params.IsValid() {
		return
	}

	// Iterate through fields and randomize adjustable parameters
	for i := 0; i < params.NumField(); i++ {
		field := params.Field(i)

		switch param := field.Interface().(type) {
		case FloatParameter:
			if param.Min != nil && param.Max > *param.Min {
				newValue := *param.Min + rand.Float64()*(param.Max-*param.Min)
				param.Value = newValue
				field.Set(reflect.ValueOf(param))
			}
		case IntParameter:
			if param.Min != nil && param.Max > *param.Min {
				newValue := *param.Min + rand.Intn(param.Max-*param.Min+1)
				param.Value = newValue
				field.Set(reflect.ValueOf(param))
			}
		case ColorParameter:
			param.Value = Color{
				R: colorPigment(rand.Intn(256)),
				G: colorPigment(rand.Intn(256)),
				B: colorPigment(rand.Intn(256)),
			}
			field.Set(reflect.ValueOf(param))
		case BooleanParameter:
			param.Value = rand.Float32() < 0.5
			field.Set(reflect.ValueOf(param))
		}
	}
}

func (p *RandomPattern) GetName() string {
	return "random"
}

func (p *RandomPattern) Start() {
	p.isActive = true
	p.lastPatternSwitch = time.Now()
	p.switchToRandomPattern()
}

func (p *RandomPattern) Stop() {
	p.isActive = false
	p.currentPattern = nil
}

type RandomUpdateRequest struct {
	Parameters RandomParameters `json:"parameters"`
}

func (r *RandomUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *RandomPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RandomUpdateRequest{
		Parameters: p.Parameters,
	}
}
