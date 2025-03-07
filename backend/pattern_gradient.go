package main

import (
	"errors"
	"fmt"
	"math"
)

type GradientPattern struct {
	BasePattern
	pixelMap     *PixelMap
	Parameters   GradientParameters `json:"parameters"`
	Label        string             `json:"label,omitempty"`
	currentAngle float64
}

func (p *GradientPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(GradientParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Color1.Update(newParams.Color1.Value)
	p.Parameters.Color2.Update(newParams.Color2.Value)
	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	p.Parameters.BlendSize.Update(newParams.BlendSize.Value)

	return nil
}

type GradientParameters struct {
	Color1    ColorParameter   `json:"color1"`
	Color2    ColorParameter   `json:"color2"`
	Speed     FloatParameter   `json:"speed"`
	Reversed  BooleanParameter `json:"reversed"`
	BlendSize FloatParameter   `json:"blendSize"`
}

func (p *GradientPattern) Update() {
	color1 := p.Parameters.Color1.Value
	color2 := p.Parameters.Color2.Value
	speed := p.Parameters.Speed.Value
	reversed := p.Parameters.Reversed.Value
	blendSize := p.Parameters.BlendSize.Value

	for i, pixel := range *p.pixelMap.pixels {
		calculatedColor := GetColorAtPointWithBlendSize(Point{pixel.x, pixel.y}, color1, color2, p.currentAngle, blendSize)
		(*p.pixelMap.pixels)[i].color = Color{
			R: calculatedColor.R,
			G: calculatedColor.G,
			B: calculatedColor.B,
			W: 0,
		}
	}
	if reversed {
		p.currentAngle += speed
	} else {
		p.currentAngle = (p.currentAngle - speed) + MAX_DEGREES
	}
	p.currentAngle = math.Mod(p.currentAngle, MAX_DEGREES)
}

func (p *GradientPattern) GetName() string {
	return "gradient"
}

type GradientUpdateRequest struct {
	Parameters GradientParameters `json:"parameters"`
}

func (r *GradientUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *GradientPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &GradientUpdateRequest{
		Parameters: p.Parameters,
	}
}

func GetColorAtPoint(p Point, color1 Color, color2 Color, angleDegrees float64) Color {
	angleRad := degreesToRadians(angleDegrees)

	// create a unit vector in the direction of the gradient
	gradientDir := struct{ X, Y float64 }{
		X: math.Cos(angleRad),
		Y: math.Sin(angleRad),
	}

	// calculate the maximum possible projection based on the angle
	// this finds the projection of the corner point (MAX_X,MAX_Y) onto the gradient direction
	maxProjection := math.Max(
		math.Max(
			projectPoint(Point{MIN_X, MIN_Y}, gradientDir),
			projectPoint(Point{MAX_X, MIN_Y}, gradientDir),
		),
		math.Max(
			projectPoint(Point{MIN_X, MAX_Y}, gradientDir),
			projectPoint(Point{MAX_X, MAX_Y}, gradientDir),
		),
	)

	minProjection := math.Min(
		math.Min(
			projectPoint(Point{MIN_X, MIN_Y}, gradientDir),
			projectPoint(Point{MAX_X, MIN_Y}, gradientDir),
		),
		math.Min(
			projectPoint(Point{MIN_X, MAX_Y}, gradientDir),
			projectPoint(Point{MAX_X, MAX_Y}, gradientDir),
		),
	)

	// project the current point onto the gradient direction
	projection := projectPoint(p, gradientDir)

	// normalize the projection to get a value between 0 and 1
	t := (projection - minProjection) / (maxProjection - minProjection)
	t = math.Max(0, math.Min(1, t))

	// convert RGB to HSV for better interpolation
	r1, g1, b1 := float64(color1.R)/255.0, float64(color1.G)/255.0, float64(color1.B)/255.0
	r2, g2, b2 := float64(color2.R)/255.0, float64(color2.G)/255.0, float64(color2.B)/255.0

	h1, s1, v1 := RGBtoHSV(r1, g1, b1)
	h2, s2, v2 := RGBtoHSV(r2, g2, b2)

	// boost saturation - make colors more vibrant
	s1 = math.Min(1.0, s1*1.5)
	s2 = math.Min(1.0, s2*1.5)

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

	// use non-linear blending for saturation to prevent washed-out colors
	sBlended := math.Sqrt(s1*s1*(1-t) + s2*s2*t)

	// boost the final saturation again
	sBlended = math.Min(1.0, sBlended*1.2)

	// use non-linear blending for value as well
	vBlended := math.Sqrt(v1*v1*(1-t) + v2*v2*t)

	// convert back to RGB
	rBlended, gBlended, bBlended := HSVtoRGB(hBlended, sBlended, vBlended)

	// apply gamma correction to make colors more vivid
	gamma := 0.8

	return Color{
		R: colorPigment(math.Pow(rBlended, gamma) * 255),
		G: colorPigment(math.Pow(gBlended, gamma) * 255),
		B: colorPigment(math.Pow(bBlended, gamma) * 255),
		W: 0,
	}
}

// projects a point onto a direction vector
func projectPoint(p Point, dir struct{ X, Y float64 }) float64 {
	fx, fy := float64(p.X), float64(p.Y)
	return fx*dir.X + fy*dir.Y
}

func (p *GradientPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}

func GetColorAtPointWithBlendSize(p Point, color1 Color, color2 Color, angleDegrees float64, blendSize float64) Color {
	angleRad := degreesToRadians(angleDegrees)

	// create a unit vector in the direction of the gradient
	gradientDir := struct{ X, Y float64 }{
		X: math.Cos(angleRad),
		Y: math.Sin(angleRad),
	}

	// calculate the maximum possible projection based on the angle
	maxProjection := math.Max(
		math.Max(
			projectPoint(Point{MIN_X, MIN_Y}, gradientDir),
			projectPoint(Point{MAX_X, MIN_Y}, gradientDir),
		),
		math.Max(
			projectPoint(Point{MIN_X, MAX_Y}, gradientDir),
			projectPoint(Point{MAX_X, MAX_Y}, gradientDir),
		),
	)

	minProjection := math.Min(
		math.Min(
			projectPoint(Point{MIN_X, MIN_Y}, gradientDir),
			projectPoint(Point{MAX_X, MIN_Y}, gradientDir),
		),
		math.Min(
			projectPoint(Point{MIN_X, MAX_Y}, gradientDir),
			projectPoint(Point{MAX_X, MAX_Y}, gradientDir),
		),
	)

	// project the current point onto the gradient direction
	projection := projectPoint(p, gradientDir)

	// normalize the projection to get a value between 0 and 1
	t := (projection - minProjection) / (maxProjection - minProjection)

	// blendSize of 0 means sharp transition (no blend)
	// blendSize of 1 means full gradient (maximum blend)
	if blendSize < 1.0 {
		center := 0.5
		halfBlendWidth := blendSize / 2.0

		// adjust t based on blend size
		if t < center-halfBlendWidth {
			// before blend region - use color1
			t = 0
		} else if t > center+halfBlendWidth {
			// after blend region - use color2
			t = 1
		} else {
			// inside blend region - remap t to [0,1] within this region
			t = (t - (center - halfBlendWidth)) / (2 * halfBlendWidth)
		}
	}

	t = math.Max(0, math.Min(1, t))

	// convert RGB to HSV for better interpolation
	r1, g1, b1 := float64(color1.R)/255.0, float64(color1.G)/255.0, float64(color1.B)/255.0
	r2, g2, b2 := float64(color2.R)/255.0, float64(color2.G)/255.0, float64(color2.B)/255.0

	h1, s1, v1 := RGBtoHSV(r1, g1, b1)
	h2, s2, v2 := RGBtoHSV(r2, g2, b2)

	// boost saturation - make colors more vibrant
	s1 = math.Min(1.0, s1*1.5)
	s2 = math.Min(1.0, s2*1.5)

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

	// use non-linear blending for saturation to prevent washed-out colors
	sBlended := math.Sqrt(s1*s1*(1-t) + s2*s2*t)

	// boost the final saturation again
	sBlended = math.Min(1.0, sBlended*1.2)

	// use non-linear blending for value as well
	vBlended := math.Sqrt(v1*v1*(1-t) + v2*v2*t)

	// convert back to RGB
	rBlended, gBlended, bBlended := HSVtoRGB(hBlended, sBlended, vBlended)

	// apply gamma correction to make colors more vivid
	gamma := 0.8

	return Color{
		R: colorPigment(math.Pow(rBlended, gamma) * 255),
		G: colorPigment(math.Pow(gBlended, gamma) * 255),
		B: colorPigment(math.Pow(bBlended, gamma) * 255),
		W: 0,
	}
}
