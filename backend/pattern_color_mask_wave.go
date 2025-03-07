package main

import (
	"fmt"
	"math"
	"time"
)

type WaveColorMask struct {
	BasePattern
	Parameters WaveParameters `json:"parameters"`
	Label      string         `json:"label,omitempty"`
	startTime  time.Time
}

type WaveParameters struct {
	Color1           ColorParameter `json:"color1"`
	Color2           ColorParameter `json:"color2"`
	WaveSpeed        FloatParameter `json:"waveSpeed"`
	WaveFrequency    FloatParameter `json:"waveFrequency"`
	WaveCount        IntParameter   `json:"waveCount"`
	WaveDirection    FloatParameter `json:"waveDirection"`
	InterferenceMode IntParameter   `json:"interferenceMode"`
	Amplitude        FloatParameter `json:"amplitude"`
}

func (p *WaveColorMask) GetColorAt(point Point) Color {
	if p.startTime.IsZero() {
		p.startTime = time.Now()
	}

	elapsed := time.Since(p.startTime).Seconds()

	// get parameters
	color1 := p.Parameters.Color1.Value
	color2 := p.Parameters.Color2.Value
	waveSpeed := p.Parameters.WaveSpeed.Value
	waveFrequency := p.Parameters.WaveFrequency.Value
	waveCount := p.Parameters.WaveCount.Value
	waveDirection := p.Parameters.WaveDirection.Value * (math.Pi / 180.0) // convert to radians
	interferenceMode := p.Parameters.InterferenceMode.Value
	amplitude := p.Parameters.Amplitude.Value

	// normalize coordinates
	nx := float64(point.X-MIN_X) / float64(MAX_X-MIN_X)
	ny := float64(point.Y-MIN_Y) / float64(MAX_Y-MIN_Y)

	// calculate wave value based on interference mode
	var waveValue float64

	switch interferenceMode {
	case 0: // linear waves
		// direction vector
		dx := math.Cos(waveDirection)
		dy := math.Sin(waveDirection)

		// project point onto direction
		projection := nx*dx + ny*dy

		// calculate wave
		waveValue = math.Sin(projection*waveFrequency*2*math.Pi + elapsed*waveSpeed)

	case 1: // radial waves
		// multiple wave sources
		waveValue = 0
		for i := 0; i < int(waveCount); i++ {
			// calculate wave source position (evenly distributed)
			angle := float64(i) * (2 * math.Pi / float64(waveCount))
			sourceX := 0.5 + 0.3*math.Cos(angle)
			sourceY := 0.5 + 0.3*math.Sin(angle)

			// distance from point to source
			distance := math.Sqrt(math.Pow(nx-sourceX, 2) + math.Pow(ny-sourceY, 2))

			// add wave contribution
			phase := distance*waveFrequency*2*math.Pi - elapsed*waveSpeed
			waveValue += math.Sin(phase) / float64(waveCount)
		}

	case 2: // spiral waves
		// distance from center
		dx := nx - 0.5
		dy := ny - 0.5
		distance := math.Sqrt(dx*dx + dy*dy)

		// angle
		angle := math.Atan2(dy, dx)
		if angle < 0 {
			angle += 2 * math.Pi
		}

		// spiral wave
		waveValue = math.Sin(angle*float64(waveCount) + distance*waveFrequency*10 - elapsed*waveSpeed)
	}

	// scale wave value to [0,1] range
	t := (waveValue*amplitude + 1) / 2
	t = math.Max(0, math.Min(1, t))

	// convert RGB to HSV for better interpolation
	r1, g1, b1 := float64(color1.R)/255.0, float64(color1.G)/255.0, float64(color1.B)/255.0
	r2, g2, b2 := float64(color2.R)/255.0, float64(color2.G)/255.0, float64(color2.B)/255.0

	h1, s1, v1 := RGBtoHSV(r1, g1, b1)
	h2, s2, v2 := RGBtoHSV(r2, g2, b2)

	// handle hue wrapping for shortest path
	if h2-h1 > 180 {
		h1 += 360
	} else if h1-h2 > 180 {
		h2 += 360
	}

	// interpolate in HSV space
	hBlended := h1*(1-t) + h2*t
	if hBlended >= 360 {
		hBlended -= 360
	}

	// use non-linear blending for saturation and value
	sBlended := math.Sqrt(s1*s1*(1-t) + s2*s2*t)
	vBlended := math.Sqrt(v1*v1*(1-t) + v2*v2*t)

	// convert back to RGB
	rBlended, gBlended, bBlended := HSVtoRGB(hBlended, sBlended, vBlended)

	// apply gamma correction
	gamma := 0.8
	return Color{
		R: colorPigment(math.Pow(rBlended, gamma) * 255),
		G: colorPigment(math.Pow(gBlended, gamma) * 255),
		B: colorPigment(math.Pow(bBlended, gamma) * 255),
		W: 0,
	}
}

func (p *WaveColorMask) Update() {
	if p.startTime.IsZero() {
		p.startTime = time.Now()
	}
}

func (p *WaveColorMask) GetName() string {
	return "waveColorMask"
}

type WaveColorMaskUpdateRequest struct {
	Parameters WaveParameters `json:"parameters"`
}

func (r *WaveColorMaskUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *WaveColorMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return &WaveColorMaskUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *WaveColorMask) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, nil)
}

func (p *WaveColorMask) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(WaveParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for WaveColorMask")
	}

	p.Parameters.Color1.Update(newParams.Color1.Value)
	p.Parameters.Color2.Update(newParams.Color2.Value)
	p.Parameters.WaveSpeed.Update(newParams.WaveSpeed.Value)
	p.Parameters.WaveFrequency.Update(newParams.WaveFrequency.Value)
	p.Parameters.WaveCount.Update(newParams.WaveCount.Value)
	p.Parameters.WaveDirection.Update(newParams.WaveDirection.Value)
	p.Parameters.InterferenceMode.Update(newParams.InterferenceMode.Value)
	p.Parameters.Amplitude.Update(newParams.Amplitude.Value)

	return nil
}
