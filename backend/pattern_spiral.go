package main

import (
	"errors"
	"fmt"
	"math"
)

type SpiralPattern struct {
	BasePattern
	pixelMap        *PixelMap
	currentRotation float64
	Parameters      SpiralParameters `json:"parameters"`
	Label           string           `json:"label,omitempty"`
}

func (p *SpiralPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(SpiralParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Color1.Update(newParams.Color1.Value)
	p.Parameters.Color2.Update(newParams.Color2.Value)
	p.Parameters.MaxTurns.Update(newParams.MaxTurns.Value)
	p.Parameters.Width.Update(newParams.Width.Value)
	return nil
}

type SpiralParameters struct {
	Speed    FloatParameter `json:"speed"`
	Color1   ColorParameter `json:"color1"`
	Color2   ColorParameter `json:"color2"`
	MaxTurns IntParameter   `json:"maxTurns"`
	Width    FloatParameter `json:"width"`
}

func (p *SpiralPattern) Update() {
	color1 := p.Parameters.Color1.Value
	color2 := p.Parameters.Color2.Value
	speed := p.Parameters.Speed.Value
	width := p.Parameters.Width.Value

	params := SpiralParams{
		Radius:       0,
		Width:        width,
		MaxTurns:     float64(p.Parameters.MaxTurns.Value),
		Rotation:     p.currentRotation,
		Center:       Point{400, 400},
		QuadrantSize: 800,
	}

	for i, pixel := range *p.pixelMap.pixels {
		point := Point{pixel.x, pixel.y}
		if isPointBetweenSpirals(point, params) {
			if p.GetColorMask() != nil {
				(*p.pixelMap.pixels)[i].color = p.GetColorMask().GetColorAt(point)
			} else {
				(*p.pixelMap.pixels)[i].color = color1
			}
		} else {
			(*p.pixelMap.pixels)[i].color = color2
		}
	}
	p.currentRotation += speed

	p.currentRotation = math.Mod(p.currentRotation, MAX_DEGREES)
}

func (p *SpiralPattern) GetName() string {
	return "spiral"
}

type SpiralUpdateRequest struct {
	Parameters SpiralParameters `json:"parameters"`
}

func (r *SpiralUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *SpiralPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &SpiralUpdateRequest{
		Parameters: p.Parameters,
	}
}

// contains all parameters needed to define our spiral
type SpiralParams struct {
	Radius       float64 // starting radius
	Width        float64 // width of the spiral
	MaxTurns     float64 // maximum number of turns
	Rotation     float64 // angle
	Center       Point   // center of the spiral
	QuadrantSize float64 // size of the quadrant
}

// calculate the growth rate (b) needed to fill the window for given turns
func calculateGrowthRate(params SpiralParams) float64 {
	maxRadius := math.Sqrt(2) * 400 // half of 800x800 window
	maxTheta := params.MaxTurns * 2 * math.Pi
	return (maxRadius - params.Radius) / maxTheta
}

// convert from cartesian to polar coordinates
func toPolar(p Point, center Point, rotationRad float64) (r, theta float64) {
	dx := float64(p.X - center.X)
	dy := float64(p.Y - center.Y)
	r = math.Sqrt(dx*dx + dy*dy)

	// get base angle and adjust for rotation
	theta = math.Atan2(dy, dx) - rotationRad

	// normalize to [0, 2π]
	if theta < 0 {
		theta += 2 * math.Pi
	}
	return r, theta
}

// find which turn of the spiral is closest to the given point
func findClosestTurn(r float64, baseTheta float64, params SpiralParams, b float64) float64 {
	minDist := math.MaxFloat64
	bestTheta := 0.0

	// estimate which turn we might be on
	estimatedTurn := (r - params.Radius) / (b * 2 * math.Pi)

	// check a few turns around our estimate
	startTurn := math.Max(0, math.Floor(estimatedTurn-1))
	endTurn := math.Min(params.MaxTurns, math.Ceil(estimatedTurn+1))

	for turn := startTurn; turn <= endTurn; turn++ {
		theta := baseTheta + turn*2*math.Pi
		expectedR := params.Radius + b*theta
		dist := math.Abs(r - expectedR)

		if dist < minDist {
			minDist = dist
			bestTheta = theta
		}
	}

	return bestTheta
}

// Check if a point lies between the two spirals
func isPointBetweenSpirals(p Point, params SpiralParams) bool {
	// calculate growth rate based on turns
	b := calculateGrowthRate(params)

	// Convert rotation from degrees to radians
	rotationRad := params.Rotation * math.Pi / 180

	// get initial polar coordinates with rotation
	r, baseTheta := toPolar(p, params.Center, rotationRad)

	// find the actual theta considering multiple turns
	theta := findClosestTurn(r, baseTheta, params, b)

	// check if theta is within our maximum turns
	if theta > params.MaxTurns*2*math.Pi {
		return false
	}

	// calculate the expected radius for this angle
	expectedR := params.Radius + b*theta

	// check if point is between inner and outer spiral
	innerR := expectedR - params.Width/2
	outerR := expectedR + params.Width/2

	return r >= innerR && r <= outerR
}

func (p *SpiralPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
