package main

import (
	"errors"
	"fmt"
	"math"
)

type GradientPattern struct {
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

	return nil
}

type GradientParameters struct {
	Color1   ColorParameter   `json:"color1"`
	Color2   ColorParameter   `json:"color2"`
	Speed    FloatParameter   `json:"speed"`
	Reversed BooleanParameter `json:"reversed"`
}

func (p *GradientPattern) Update() {
	color1 := p.Parameters.Color1.Value
	color2 := p.Parameters.Color2.Value
	speed := p.Parameters.Speed.Value
	reversed := p.Parameters.Reversed.Value

	for i, pixel := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = GetColorAtPoint(Point{pixel.x, pixel.y}, color1, color2, p.currentAngle)
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
		Parameters: GradientParameters{},
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

	// interpolate between the two colors
	return Color{
		R: colorPigment(float64(color1.R)*(1-t) + float64(color2.R)*t),
		G: colorPigment(float64(color1.G)*(1-t) + float64(color2.G)*t),
		B: colorPigment(float64(color1.B)*(1-t) + float64(color2.B)*t),
	}
}

// projects a point onto a direction vector
func projectPoint(p Point, dir struct{ X, Y float64 }) float64 {
	fx, fy := float64(p.X), float64(p.Y)
	return fx*dir.X + fy*dir.Y
}
