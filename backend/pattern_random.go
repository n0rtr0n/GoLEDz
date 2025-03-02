package main

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

var ErrInvalidParameters = fmt.Errorf("invalid parameters")

type RandomPattern struct {
	BasePattern
	pixelMap            *PixelMap
	patterns            map[string]Pattern
	currentPattern      Pattern
	nextPattern         Pattern
	lastSwitchTime      time.Time
	transitionStartTime time.Time
	inTransition        bool
	Parameters          RandomPatternParameters `json:"parameters"`
}

type RandomPatternParameters struct {
	SwitchInterval      FloatParameter   `json:"switchInterval"`
	RandomizeColorMasks BooleanParameter `json:"randomizeColorMasks"`
	TransitionTime      FloatParameter   `json:"transitionTime"`
}

func (p *RandomPattern) Update() {
	if p.inTransition {
		elapsed := time.Since(p.transitionStartTime).Seconds()
		transitionDuration := p.Parameters.TransitionTime.Value
		progress := float64(elapsed) / transitionDuration

		if progress >= 1.0 {
			p.inTransition = false
			p.currentPattern = p.nextPattern
			p.nextPattern = nil
			p.lastSwitchTime = time.Now()
		} else {
			if p.currentPattern != nil && p.nextPattern != nil {
				currentPixels := make([]Pixel, len(*p.pixelMap.pixels))
				copy(currentPixels, *p.pixelMap.pixels)

				p.currentPattern.Update()
				sourcePixels := make([]Pixel, len(*p.pixelMap.pixels))
				copy(sourcePixels, *p.pixelMap.pixels)

				copy(*p.pixelMap.pixels, currentPixels)

				p.nextPattern.Update()
				targetPixels := make([]Pixel, len(*p.pixelMap.pixels))
				copy(targetPixels, *p.pixelMap.pixels)

				for i := range *p.pixelMap.pixels {
					(*p.pixelMap.pixels)[i].color = blendColors(
						sourcePixels[i].color,
						targetPixels[i].color,
						progress)
				}
				return
			}
		}
	}

	if !p.inTransition && time.Since(p.lastSwitchTime).Seconds() > p.Parameters.SwitchInterval.Value {
		p.startTransition()
	}

	if p.currentPattern != nil {
		p.currentPattern.Update()

		for i, pixel := range *p.pixelMap.pixels {
			(*p.pixelMap.pixels)[i].color = pixel.color
		}
	}
}

func (p *RandomPattern) startTransition() {
	p.selectRandomPattern()

	if p.nextPattern != nil {
		if p.Parameters.RandomizeColorMasks.Value {
			p.selectRandomColorMask()
		} else if p.GetColorMask() != nil {
			p.nextPattern.SetColorMask(p.GetColorMask())
		}

		p.inTransition = true
		p.transitionStartTime = time.Now()
	}
}

func (p *RandomPattern) selectRandomPattern() {
	var patternNames []string
	for name, _ := range p.patterns {
		if name != "random" && (p.currentPattern == nil || name != p.currentPattern.GetName()) {
			patternNames = append(patternNames, name)
		}
	}

	if len(patternNames) == 0 {
		return
	}

	nextPatternName := patternNames[rand.Intn(len(patternNames))]
	p.nextPattern = p.patterns[nextPatternName]

	p.randomizeParameters(p.nextPattern)
}

func (p *RandomPattern) selectRandomColorMask() {
	colorMasks := registerColorMasks()
	if len(colorMasks) > 0 {
		var maskNames []string
		for name := range colorMasks {
			maskNames = append(maskNames, name)
		}

		randomMaskName := maskNames[rand.Intn(len(maskNames))]
		randomMask := colorMasks[randomMaskName]

		p.randomizeParameters(randomMask)

		p.nextPattern.SetColorMask(randomMask)
	}
}

func (p *RandomPattern) GetName() string {
	return "random"
}

func (p *RandomPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RandomPatternParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.SwitchInterval.Update(newParams.SwitchInterval.Value)
	p.Parameters.RandomizeColorMasks.Update(newParams.RandomizeColorMasks.Value)
	p.Parameters.TransitionTime.Update(newParams.TransitionTime.Value)
	return nil
}

func (p *RandomPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RandomPatternUpdateRequest{
		Parameters: p.Parameters,
	}
}

type RandomPatternUpdateRequest struct {
	Parameters RandomPatternParameters `json:"parameters"`
}

func (r *RandomPatternUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *RandomPattern) TransitionFrom(source Pattern, progress float64) {
	if progress < 1.0 {
		return
	}
	if p.currentPattern != nil {
		p.startTransition()
		return
	}
	p.selectRandomPattern()
	if p.nextPattern != nil {
		p.currentPattern = p.nextPattern
		p.nextPattern = nil

		// if randomizing color masks is enabled, select a random mask
		if p.Parameters.RandomizeColorMasks.Value {
			colorMasks := registerColorMasks()
			if len(colorMasks) > 0 {
				var maskNames []string
				for name := range colorMasks {
					maskNames = append(maskNames, name)
				}

				randomMaskName := maskNames[rand.Intn(len(maskNames))]
				randomMask := colorMasks[randomMaskName]

				// randomize the mask's parameters
				p.randomizeParameters(randomMask)

				// set the mask on the current pattern
				p.currentPattern.SetColorMask(randomMask)
				p.SetColorMask(randomMask)
			}
		}

		p.currentPattern.Update()

		// copy colors to this pattern
		for i, pixel := range *p.pixelMap.pixels {
			(*p.pixelMap.pixels)[i].color = pixel.color
		}

		p.lastSwitchTime = time.Now()
	}
}

func (p *RandomPattern) randomizeParameters(pattern Pattern) {
	// use reflection to access the Parameters field of the pattern
	patternValue := reflect.ValueOf(pattern).Elem()
	paramsField := patternValue.FieldByName("Parameters")

	if !paramsField.IsValid() {
		fmt.Printf("Pattern %s doesn't have Parameters field\n", pattern.GetName())
		return
	}

	for i := range paramsField.NumField() {
		field := paramsField.Field(i)
		fieldType := paramsField.Type().Field(i)

		if !field.CanInterface() {
			continue
		}

		if param, ok := field.Addr().Interface().(Parameter); ok {
			param.Randomize()
		} else {
			fmt.Printf("  Field %s doesn't implement Parameter interface\n", fieldType.Name)
		}
	}
}
