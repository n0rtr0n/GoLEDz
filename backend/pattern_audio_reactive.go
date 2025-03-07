package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type AudioReactivePattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters AudioReactiveParameters `json:"parameters"`
	Label      string                  `json:"label,omitempty"`
	lastUpdate time.Time
	phase      float64
}

type AudioReactiveParameters struct {
	Sensitivity   FloatParameter `json:"sensitivity"`
	ColorSpeed    FloatParameter `json:"colorSpeed"`
	BaseColor     ColorParameter `json:"baseColor"`
	AccentColor   ColorParameter `json:"accentColor"`
	EffectType    IntParameter   `json:"effectType"` // 0=pulse, 1=wave, 2=sparkle
	SmoothingTime FloatParameter `json:"smoothingTime"`
}

func (p *AudioReactivePattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(AudioReactiveParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Sensitivity.Update(newParams.Sensitivity.Value)
	p.Parameters.ColorSpeed.Update(newParams.ColorSpeed.Value)
	p.Parameters.BaseColor.Update(newParams.BaseColor.Value)
	p.Parameters.AccentColor.Update(newParams.AccentColor.Value)
	p.Parameters.EffectType.Update(newParams.EffectType.Value)
	p.Parameters.SmoothingTime.Update(newParams.SmoothingTime.Value)
	return nil
}

func (p *AudioReactivePattern) Update() {
	// Initialize if this is the first update
	if p.lastUpdate.IsZero() {
		p.lastUpdate = time.Now()
		p.phase = 0
	}

	// Calculate time delta
	now := time.Now()
	deltaTime := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now

	// Update phase for color cycling
	p.phase += deltaTime * p.Parameters.ColorSpeed.Value
	if p.phase > 360 {
		p.phase -= 360
	}

	// Get audio level (this would need to be implemented based on your audio input method)
	audioLevel := p.getAudioLevel() * p.Parameters.Sensitivity.Value

	// Apply the selected effect
	switch p.Parameters.EffectType.Value {
	case 0:
		p.applyPulseEffect(audioLevel)
	case 1:
		p.applyWaveEffect(audioLevel)
	case 2:
		p.applySparkleEffect(audioLevel)
	}

	// Update the color mask if we have one
	if p.GetColorMask() != nil {
		p.GetColorMask().Update()
	}
}

// This is a placeholder - you would need to implement actual audio input
func (p *AudioReactivePattern) getAudioLevel() float64 {
	// Placeholder for actual audio input
	// In a real implementation, this would get audio levels from a microphone
	// For now, we'll simulate with a sine wave
	return 0.5 + 0.5*math.Sin(p.phase*0.1)
}

func (p *AudioReactivePattern) applyPulseEffect(audioLevel float64) {
	baseColor := p.Parameters.BaseColor.Value
	accentColor := p.Parameters.AccentColor.Value

	// Blend between base and accent color based on audio level
	for i := range *p.pixelMap.pixels {
		// Apply color mask if available
		if p.GetColorMask() != nil {
			pixel := (*p.pixelMap.pixels)[i]
			maskColor := p.GetColorMask().GetColorAt(Point{pixel.x, pixel.y})

			// Blend between base color and mask color based on audio level
			(*p.pixelMap.pixels)[i].color = Color{
				R: colorPigment(float64(baseColor.R)*(1-audioLevel) + float64(maskColor.R)*audioLevel),
				G: colorPigment(float64(baseColor.G)*(1-audioLevel) + float64(maskColor.G)*audioLevel),
				B: colorPigment(float64(baseColor.B)*(1-audioLevel) + float64(maskColor.B)*audioLevel),
				W: 0,
			}
		} else {
			// Blend between base and accent color based on audio level
			(*p.pixelMap.pixels)[i].color = Color{
				R: colorPigment(float64(baseColor.R)*(1-audioLevel) + float64(accentColor.R)*audioLevel),
				G: colorPigment(float64(baseColor.G)*(1-audioLevel) + float64(accentColor.G)*audioLevel),
				B: colorPigment(float64(baseColor.B)*(1-audioLevel) + float64(accentColor.B)*audioLevel),
				W: 0,
			}
		}
	}
}

func (p *AudioReactivePattern) applyWaveEffect(audioLevel float64) {
	baseColor := p.Parameters.BaseColor.Value
	accentColor := p.Parameters.AccentColor.Value

	// Find max Y value
	maxY := int16(0)
	for _, pixel := range *p.pixelMap.pixels {
		if pixel.y > maxY {
			maxY = pixel.y
		}
	}

	// Create a wave that moves up based on audio level
	waveHeight := float64(maxY) * audioLevel

	for i, pixel := range *p.pixelMap.pixels {
		// Calculate distance from wave
		distFromWave := math.Abs(float64(pixel.y) - waveHeight)
		waveEffect := math.Max(0, 1-distFromWave/30.0) // 30 pixel fade distance

		// Apply color mask if available
		if p.GetColorMask() != nil {
			maskColor := p.GetColorMask().GetColorAt(Point{pixel.x, pixel.y})

			// Blend between base color and mask color based on wave effect
			(*p.pixelMap.pixels)[i].color = Color{
				R: colorPigment(float64(baseColor.R)*(1-waveEffect) + float64(maskColor.R)*waveEffect),
				G: colorPigment(float64(baseColor.G)*(1-waveEffect) + float64(maskColor.G)*waveEffect),
				B: colorPigment(float64(baseColor.B)*(1-waveEffect) + float64(maskColor.B)*waveEffect),
				W: 0,
			}
		} else {
			// Blend between base and accent color based on wave effect
			(*p.pixelMap.pixels)[i].color = Color{
				R: colorPigment(float64(baseColor.R)*(1-waveEffect) + float64(accentColor.R)*waveEffect),
				G: colorPigment(float64(baseColor.G)*(1-waveEffect) + float64(accentColor.G)*waveEffect),
				B: colorPigment(float64(baseColor.B)*(1-waveEffect) + float64(accentColor.B)*waveEffect),
				W: 0,
			}
		}
	}
}

func (p *AudioReactivePattern) applySparkleEffect(audioLevel float64) {
	baseColor := p.Parameters.BaseColor.Value
	accentColor := p.Parameters.AccentColor.Value

	// Set all pixels to base color first
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = baseColor
	}

	// Number of sparkles based on audio level
	sparkleCount := int(audioLevel * 50)

	// Add random sparkles
	for i := 0; i < sparkleCount; i++ {
		idx := rand.Intn(len(*p.pixelMap.pixels))

		// Apply color mask if available
		if p.GetColorMask() != nil {
			pixel := (*p.pixelMap.pixels)[idx]
			maskColor := p.GetColorMask().GetColorAt(Point{pixel.x, pixel.y})
			(*p.pixelMap.pixels)[idx].color = maskColor
		} else {
			(*p.pixelMap.pixels)[idx].color = accentColor
		}
	}
}

func (p *AudioReactivePattern) GetName() string {
	return "audioReactive"
}

type AudioReactiveUpdateRequest struct {
	Parameters AudioReactiveParameters `json:"parameters"`
}

func (r *AudioReactiveUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *AudioReactivePattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &AudioReactiveUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *AudioReactivePattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
