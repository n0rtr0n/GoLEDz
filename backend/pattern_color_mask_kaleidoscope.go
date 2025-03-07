package main

import (
	"fmt"
	"math"
	"time"
)

type KaleidoscopeColorMask struct {
	BasePattern
	Parameters   KaleidoscopeParameters `json:"parameters"`
	Label        string                 `json:"label,omitempty"`
	startTime    time.Time
	currentAngle float64
}

type KaleidoscopeParameters struct {
	Color1         ColorParameter `json:"color1"`
	Color2         ColorParameter `json:"color2"`
	Color3         ColorParameter `json:"color3"`
	RotationSpeed  FloatParameter `json:"rotationSpeed"`
	Segments       IntParameter   `json:"segments"`
	ZoomLevel      FloatParameter `json:"zoomLevel"`
	Distortion     FloatParameter `json:"distortion"`
	ColorBlendMode IntParameter   `json:"colorBlendMode"`
}

func (p *KaleidoscopeColorMask) GetColorAt(point Point) Color {
	if p.startTime.IsZero() {
		p.startTime = time.Now()
	}

	// get parameters
	color1 := p.Parameters.Color1.Value
	color2 := p.Parameters.Color2.Value
	color3 := p.Parameters.Color3.Value
	segments := p.Parameters.Segments.Value
	zoomLevel := p.Parameters.ZoomLevel.Value
	distortion := p.Parameters.Distortion.Value
	colorBlendMode := p.Parameters.ColorBlendMode.Value

	// center coordinates
	centerX := (MAX_X + MIN_X) / 2
	centerY := (MAX_Y + MIN_Y) / 2

	// translate to center
	x := float64(point.X) - float64(centerX)
	y := float64(point.Y) - float64(centerY)

	// apply rotation
	rotatedX := x*math.Cos(p.currentAngle) - y*math.Sin(p.currentAngle)
	rotatedY := x*math.Sin(p.currentAngle) + y*math.Cos(p.currentAngle)

	// convert to polar coordinates
	r := math.Sqrt(rotatedX*rotatedX+rotatedY*rotatedY) * zoomLevel
	theta := math.Atan2(rotatedY, rotatedX)
	if theta < 0 {
		theta += 2 * math.Pi
	}

	// apply kaleidoscope effect by mirroring the angle
	segmentAngle := 2 * math.Pi / float64(segments)
	segmentIndex := int(theta / segmentAngle)
	segmentPosition := theta - float64(segmentIndex)*segmentAngle

	// mirror within segment
	if segmentIndex%2 == 1 {
		segmentPosition = segmentAngle - segmentPosition
	}

	// apply distortion
	distortedR := r + distortion*math.Sin(segmentPosition*5)

	// normalize coordinates for color mapping
	normalizedR := math.Min(distortedR/(MAX_X-MIN_X), 1.0)
	normalizedTheta := segmentPosition / segmentAngle

	// determine color based on blend mode
	var finalColor Color

	switch colorBlendMode {
	case 0: // radial gradient
		t := normalizedR
		finalColor = blendThreeColors(color1, color2, color3, t)

	case 1: // angular gradient
		t := normalizedTheta
		finalColor = blendThreeColors(color1, color2, color3, t)

	case 2: // complex blend
		// use both radius and angle for a more complex pattern
		t1 := normalizedR
		t2 := normalizedTheta

		pattern := math.Sin(t1*math.Pi*3 + t2*math.Pi*5)
		t := (pattern + 1) / 2 // normalize to [0,1]

		finalColor = blendThreeColors(color1, color2, color3, t)

	case 3: // time-based blend
		elapsed := time.Since(p.startTime).Seconds()
		timePhase := math.Sin(elapsed*0.5)*0.5 + 0.5

		// combine time with position
		t := (normalizedR + normalizedTheta + timePhase) / 3
		t = math.Mod(t, 1.0)

		finalColor = blendThreeColors(color1, color2, color3, t)
	}

	return finalColor
}

// helper function to blend between three colors
func blendThreeColors(c1, c2, c3 Color, t float64) Color {
	// convert to HSV for better blending
	r1, g1, b1 := float64(c1.R)/255.0, float64(c1.G)/255.0, float64(c1.B)/255.0
	r2, g2, b2 := float64(c2.R)/255.0, float64(c2.G)/255.0, float64(c2.B)/255.0
	r3, g3, b3 := float64(c3.R)/255.0, float64(c3.G)/255.0, float64(c3.B)/255.0

	h1, s1, v1 := RGBtoHSV(r1, g1, b1)
	h2, s2, v2 := RGBtoHSV(r2, g2, b2)
	h3, s3, v3 := RGBtoHSV(r3, g3, b3)

	// adjust for shortest path
	if h2-h1 > 180 {
		h1 += 360
	} else if h1-h2 > 180 {
		h2 += 360
	}

	if h3-h2 > 180 {
		h2 += 360
	} else if h2-h3 > 180 {
		h3 += 360
	}

	// determine which segment we're in
	var h, s, v float64
	if t < 0.5 {
		// blend between color1 and color2
		t2 := t * 2 // scale to [0,1]
		h = h1*(1-t2) + h2*t2
		s = s1*(1-t2) + s2*t2
		v = v1*(1-t2) + v2*t2
	} else {
		// blend between color2 and color3
		t2 := (t - 0.5) * 2 // scale to [0,1]
		h = h2*(1-t2) + h3*t2
		s = s2*(1-t2) + s3*t2
		v = v2*(1-t2) + v3*t2
	}

	// normalize hue
	if h >= 360 {
		h -= 360
	}

	// convert back to RGB
	r, g, b := HSVtoRGB(h, s, v)

	// apply gamma correction
	gamma := 0.8
	return Color{
		R: colorPigment(math.Pow(r, gamma) * 255),
		G: colorPigment(math.Pow(g, gamma) * 255),
		B: colorPigment(math.Pow(b, gamma) * 255),
		W: 0,
	}
}

func (p *KaleidoscopeColorMask) Update() {
	if p.startTime.IsZero() {
		p.startTime = time.Now()
	}

	rotationSpeed := p.Parameters.RotationSpeed.Value
	p.currentAngle += rotationSpeed * 0.01
	p.currentAngle = math.Mod(p.currentAngle, 2*math.Pi)
}

func (p *KaleidoscopeColorMask) GetName() string {
	return "kaleidoscopeColorMask"
}

type KaleidoscopeColorMaskUpdateRequest struct {
	Parameters KaleidoscopeParameters `json:"parameters"`
}

func (r *KaleidoscopeColorMaskUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *KaleidoscopeColorMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return &KaleidoscopeColorMaskUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *KaleidoscopeColorMask) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, nil)
}

func (p *KaleidoscopeColorMask) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(KaleidoscopeParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for KaleidoscopeColorMask")
	}

	p.Parameters.Color1.Update(newParams.Color1.Value)
	p.Parameters.Color2.Update(newParams.Color2.Value)
	p.Parameters.Color3.Update(newParams.Color3.Value)
	p.Parameters.RotationSpeed.Update(newParams.RotationSpeed.Value)
	p.Parameters.Segments.Update(newParams.Segments.Value)
	p.Parameters.ZoomLevel.Update(newParams.ZoomLevel.Value)
	p.Parameters.Distortion.Update(newParams.Distortion.Value)
	p.Parameters.ColorBlendMode.Update(newParams.ColorBlendMode.Value)

	return nil
}
