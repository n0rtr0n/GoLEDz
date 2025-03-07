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
	for name := range p.patterns {
		if name != "random" && name != "lightsOff" && (p.currentPattern == nil || name != p.currentPattern.GetName()) {
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
	patternType := patternValue.Type()
	patternName := patternType.Name()

	fmt.Printf("Randomizing parameters for %s\n", patternName)

	paramsField := patternValue.FieldByName("Parameters")

	if !paramsField.IsValid() {
		fmt.Printf("Pattern %s doesn't have Parameters field\n", pattern.GetName())
		return
	}

	for i := 0; i < paramsField.NumField(); i++ {
		field := paramsField.Field(i)
		fieldType := paramsField.Type().Field(i)
		fieldName := fieldType.Name

		// skip unexported fields
		if !field.CanInterface() {
			continue
		}

		switch fieldValue := field.Addr().Interface().(type) {
		case *FloatParameter:
			if fieldValue.Min != nil {
				oldValue := fieldValue.Value
				min := *fieldValue.Min
				max := fieldValue.Max
				defaultValue := fieldValue.Value

				// Calculate constrained min and max (halfway between default and actual min/max)
				constrainedMin := min + (defaultValue-min)/2
				constrainedMax := defaultValue + (max-defaultValue)/2

				// Generate random value within constrained range
				fieldValue.Value = constrainedMin + rand.Float64()*(constrainedMax-constrainedMin)

				fmt.Printf("  %s.%s: %.2f -> %.2f (range: %.2f to %.2f, constrained: %.2f to %.2f)\n",
					patternName, fieldName, oldValue, fieldValue.Value,
					min, max, constrainedMin, constrainedMax)
			}
		case *IntParameter:
			if fieldValue.Min != nil {
				oldValue := fieldValue.Value
				min := *fieldValue.Min
				max := fieldValue.Max
				defaultValue := fieldValue.Value

				// Calculate constrained min and max (halfway between default and actual min/max)
				constrainedMin := min + (defaultValue-min)/2
				constrainedMax := defaultValue + (max-defaultValue)/2

				// Generate random value within constrained range
				fieldValue.Value = min + rand.Intn(max-min+1)

				// Ensure the value is within the constrained range
				if fieldValue.Value < int(constrainedMin) {
					fieldValue.Value = int(constrainedMin)
				} else if fieldValue.Value > int(constrainedMax) {
					fieldValue.Value = int(constrainedMax)
				}

				fmt.Printf("  %s.%s: %d -> %d (range: %d to %d, constrained: %.1f to %.1f)\n",
					patternName, fieldName, oldValue, fieldValue.Value,
					min, max, constrainedMin, constrainedMax)
			}
		case *BooleanParameter:
			oldValue := fieldValue.Value
			fieldValue.Value = rand.Intn(2) == 1
			fmt.Printf("  %s.%s: %v -> %v\n", patternName, fieldName, oldValue, fieldValue.Value)
		case *ColorParameter:
			oldColor := fieldValue.Value

			// Generate a random hue (0-360), high saturation (0.7-1.0), and high value (0.7-1.0)
			h := rand.Float64() * 360
			s := 0.7 + rand.Float64()*0.3
			v := 0.7 + rand.Float64()*0.3

			// Convert HSV to RGB
			r, g, b := HSVtoRGB(h, s, v)

			fieldValue.Value = Color{
				R: colorPigment(r * 255),
				G: colorPigment(g * 255),
				B: colorPigment(b * 255),
				W: 0, // Keep W at 0
			}

			fmt.Printf("  %s.%s: RGB(%d,%d,%d) -> RGB(%d,%d,%d) (HSV: %.1f,%.1f,%.1f)\n",
				patternName, fieldName,
				oldColor.R, oldColor.G, oldColor.B,
				fieldValue.Value.R, fieldValue.Value.G, fieldValue.Value.B,
				h, s, v)
		default:
			fmt.Printf("  %s.%s type %T doesn't have randomization implemented\n",
				patternName, fieldName, fieldValue)
		}
	}
}
