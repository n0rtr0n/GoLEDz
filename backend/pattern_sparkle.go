package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
)

const MAX_SPARKLE_TTL = 80
const MAX_SPARKLES = 1000
const MAX_SPARKLE_SPEED = 2.0
const SPARKLE_STARTING_SIZE = 25.0
const SPARKLE_CHANCE_TO_CREATE = 80.0
const SPARKLE_DEFAULT_ROTATION = 45.0 // 45 degree angle, slightly rotated box

// every cycle, check to see if we are below the max number of sparkles
// if not, then have a CHANCE (1/5) at creating a new one, up to the max limit
// 50/50 chance of growing/shrinking unless ttl = size, in which case will always shrink
// ttl is set randomly (1-100) at the creation, and when the sparkle hits 0, it dies and is removed

type Sparkle struct {
	x        int
	y        int
	rotation float64
	size     float64
	speed    float64
	ttl      float64
}

type SparklePattern struct {
	BasePattern
	pixelMap   *PixelMap
	sparkles   []*Sparkle
	Parameters SparkleParameters `json:"parameters"`
	Label      string            `json:"label,omitempty"`
}

func (p *SparklePattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(SparkleParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Color.Update(newParams.Color.Value)
	return nil
}

type SparkleParameters struct {
	Color ColorParameter `json:"color"`
}

func (p *SparklePattern) Update() {
	color := p.Parameters.Color.Value

	if len(p.sparkles) < MAX_SPARKLES {
		if randomChancePercent(SPARKLE_CHANCE_TO_CREATE) {
			p.sparkles = append(p.sparkles, &Sparkle{
				x:        rand.IntN(MAX_X),
				y:        rand.IntN(MAX_Y),
				rotation: SPARKLE_DEFAULT_ROTATION,
				size:     SPARKLE_STARTING_SIZE,
				speed:    rand.Float64() * MAX_SPARKLE_SPEED,
				// this ensures ttl never dips below starting size
				ttl: SPARKLE_STARTING_SIZE + (rand.Float64() * MAX_SPARKLE_TTL),
			})
		}
	}

	for i := 0; i < len(p.sparkles); {
		sparkle := p.sparkles[i]

		if randomChancePercent(85) { // grow, never grow more than ttl
			if sparkle.size < sparkle.ttl {
				sparkle.size += sparkle.speed
			}
		} else if randomChancePercent(15) { // shrink, never dip below 1.0
			if (sparkle.size + sparkle.speed) > 1.0 {
				sparkle.size -= sparkle.speed
			}
		}
		sparkle.rotation += sparkle.speed

		// low ttl will always cause us to shrink
		if sparkle.ttl < sparkle.size {
			sparkle.size = sparkle.ttl
		}

		// remove this sparkle if we're at the end of our ttl
		if sparkle.ttl <= 0.0 {
			p.sparkles = append(p.sparkles[:i], p.sparkles[i+1:]...)
		} else {
			i++
		}
		sparkle.ttl -= 1.0
	}

	for i, pixel := range *p.pixelMap.pixels {
		point := Point{pixel.x, pixel.y}
		if pointIsBetweenAnySparkle(point, p.sparkles) {
			if p.GetColorMask() != nil {
				(*p.pixelMap.pixels)[i].color = p.GetColorMask().GetColorAt(point)
			} else {
				(*p.pixelMap.pixels)[i].color = color
			}
		} else {
			(*p.pixelMap.pixels)[i].color = Color{0, 0, 0, 0}
		}
	}
}

func (p *SparklePattern) GetName() string {
	return "sparkle"
}

type SparkleUpdateRequest struct {
	Parameters SparkleParameters `json:"parameters"`
}

func (r *SparkleUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *SparklePattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &SparkleUpdateRequest{
		Parameters: p.Parameters,
	}
}

func pointIsBetweenAnySparkle(p Point, sparkles []*Sparkle) bool {
	for _, sparkle := range sparkles {
		if pointIsBetweenSparkle(p, sparkle) {
			return true
		}
	}
	return false
}

func pointIsBetweenSparkle(p Point, sparkle *Sparkle) bool {
	// Step 1: Translate the point (x, y) relative to the center of the box (x1, y1)
	translatedX := float64(p.X - int16(sparkle.x))
	translatedY := float64(p.Y - int16(sparkle.y))

	// Step 2: Convert rotation angle to radians
	rotationRadians := degreesToRadians(sparkle.rotation)

	// Step 3: Calculate cosine and sine for the rotation matrix
	cosTheta := math.Cos(rotationRadians)
	sinTheta := math.Sin(rotationRadians)

	// Step 4: Rotate the point by the negative of the given angle (inverse rotation)
	rotatedX := cosTheta*translatedX + sinTheta*translatedY
	rotatedY := -sinTheta*translatedX + cosTheta*translatedY

	// Step 5: Check if the rotated point is within the bounds of the axis-aligned box
	halfSize := sparkle.size / 2
	return rotatedX >= -halfSize && rotatedX <= halfSize && rotatedY >= -halfSize && rotatedY <= halfSize
}

func (p *SparklePattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
