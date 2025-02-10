package main

import (
	"errors"
	"fmt"
	"math"
)

type SpiralPattern struct {
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
	p.Parameters.Color.Update(newParams.Color.Value)
	p.Parameters.MaxTurns.Update(newParams.MaxTurns.Value)
	p.Parameters.Width.Update(newParams.Width.Value)
	return nil
}

type SpiralParameters struct {
	Speed    FloatParameter `json:"speed"`
	Color    ColorParameter `json:"color"`
	MaxTurns IntParameter   `json:"maxTurns"`
	Width    FloatParameter `json:"width"`
}

func (p *SpiralPattern) Update() {
	color := p.Parameters.Color.Value
	speed := p.Parameters.Speed.Value
	width := p.Parameters.Width.Value

	// Example usage
	params := SpiralParams{
		A:            0,
		B:            30,
		Width:        width,
		MaxTurns:     float64(p.Parameters.MaxTurns.Value),
		Rotation:     p.currentRotation,
		Center:       Point{400, 400},
		QuadrantSize: 800,
	}

	// gradient := Gradient{
	// 	StartColor: Color{59, 130, 246}, // Blue
	// 	EndColor:   Color{139, 92, 246}, // Purple
	// }

	for i, pixel := range *p.pixelMap.pixels {
		point := Point{pixel.x, pixel.y}
		if isPointBetweenSpirals(point, params) {
			(*p.pixelMap.pixels)[i].color = color
		} else {
			(*p.pixelMap.pixels)[i].color = Color{0, 0, 0}
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
		Parameters: SpiralParameters{},
	}
}

// SpiralParams contains all parameters needed to define our spiral
type SpiralParams struct {
	A            float64 // Starting radius
	B            float64 // Growth rate
	Width        float64 // Width of the spiral
	MaxTurns     float64 // Maximum number of turns
	Rotation     float64
	Center       Point   // Center of the spiral
	QuadrantSize float64 // Size of the quadrant
}

// Calculate the growth rate (b) needed to fill the window for given turns
func calculateGrowthRate(params SpiralParams) float64 {
	maxRadius := math.Sqrt(2) * 400 // half of 800x800 window
	maxTheta := params.MaxTurns * 2 * math.Pi
	return (maxRadius - params.A) / maxTheta
}

// Convert from cartesian to polar coordinates
func toPolar(p Point, center Point, rotationRad float64) (r, theta float64) {
	dx := float64(p.X - center.X)
	dy := float64(p.Y - center.Y)
	r = math.Sqrt(dx*dx + dy*dy)

	// Get base angle and adjust for rotation
	theta = math.Atan2(dy, dx) - rotationRad

	// Normalize to [0, 2π]
	if theta < 0 {
		theta += 2 * math.Pi
	}
	return r, theta
}

// Find which turn of the spiral is closest to the given point
func findClosestTurn(r float64, baseTheta float64, params SpiralParams, b float64) float64 {
	minDist := math.MaxFloat64
	bestTheta := 0.0

	// Estimate which turn we might be on
	estimatedTurn := (r - params.A) / (b * 2 * math.Pi)

	// Check a few turns around our estimate
	startTurn := math.Max(0, math.Floor(estimatedTurn-1))
	endTurn := math.Min(params.MaxTurns, math.Ceil(estimatedTurn+1))

	for turn := startTurn; turn <= endTurn; turn++ {
		theta := baseTheta + turn*2*math.Pi
		expectedR := params.A + b*theta
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
	// Calculate growth rate based on turns
	b := calculateGrowthRate(params)

	// Convert rotation from degrees to radians
	rotationRad := params.Rotation * math.Pi / 180

	// Get initial polar coordinates with rotation
	r, baseTheta := toPolar(p, params.Center, rotationRad)

	// Find the actual theta considering multiple turns
	theta := findClosestTurn(r, baseTheta, params, b)

	// Check if theta is within our maximum turns
	if theta > params.MaxTurns*2*math.Pi {
		return false
	}

	// Calculate the expected radius for this angle
	expectedR := params.A + b*theta

	// Check if point is between inner and outer spiral
	innerR := expectedR - params.Width/2
	outerR := expectedR + params.Width/2

	return r >= innerR && r <= outerR
}
