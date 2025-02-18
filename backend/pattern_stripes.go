package main

import (
	"errors"
	"fmt"
	"math"
)

const STRIPES_STARTING_POSITION = 0
const STRIPES_ENDING_POSITION = MAX_X

type StripesPattern struct {
	pixelMap        *PixelMap
	currentPosition float64
	Parameters      StripesParameters `json:"parameters"`
	Label           string            `json:"label,omitempty"`
}

func (p *StripesPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(StripesParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Size.Update(newParams.Size.Value)
	p.Parameters.Color.Update(newParams.Color.Value)
	p.Parameters.Rotation.Update(newParams.Rotation.Value)
	p.Parameters.Stripes.Update(newParams.Stripes.Value)
	return nil
}

type StripesParameters struct {
	Speed    FloatParameter `json:"speed"`
	Size     FloatParameter `json:"size"`
	Color    ColorParameter `json:"color"`
	Rotation FloatParameter `json:"rotation"`
	Stripes  IntParameter   `json:"stripes"`
}

func (p *StripesPattern) Update() {
	speed := p.Parameters.Speed.Value
	size := p.Parameters.Size.Value
	color := p.Parameters.Color.Value
	rotation := p.Parameters.Rotation.Value
	stripes := p.Parameters.Stripes.Value

	maxPosition := float64(MAX_X)
	spaceBetweenStripes := maxPosition / float64(stripes)
	positions := []float64{}
	for i := 0; i < stripes; i++ {
		position := p.currentPosition + (float64(i) * spaceBetweenStripes)
		position = math.Mod(position, maxPosition)
		positions = append(positions, position)
	}

	for i, pixel := range *p.pixelMap.pixels {

		if isInAnyBox(Point{pixel.x, pixel.y}, size, rotation, positions) {
			(*p.pixelMap.pixels)[i].color = color
		} else {
			(*p.pixelMap.pixels)[i].color = Color{}
		}
	}

	p.currentPosition = math.Mod(p.currentPosition+speed, float64(maxPosition))
}
func isInAnyBox(point Point, size float64, rotation float64, positions []float64) bool {
	for _, position := range positions {
		if isInBox(point, size, rotation, position) {
			return true
		}
	}

	return false
}

func isInBox(point Point, size float64, rotation float64, position float64) bool {

	// the rotation will impact the starting position. we want to start approximately 2x away from the center
	// and always move towards the center
	rotationCenterX := float64(CENTER_X)
	rotationCenterY := float64(CENTER_Y)
	boxCenterX := STRIPES_STARTING_POSITION + position
	boxCenterY := float64(CENTER_Y)
	boxWidth := size
	boxHeight := float64(MAX_Y)

	angleInRadians := degreesToRadians(rotation) / 2

	// translate point to rotation center coordinate system
	translatedX := float64(point.X) - rotationCenterX
	translatedY := float64(point.Y) - rotationCenterY

	// apply inverse rotation
	cosAngle := math.Cos(-angleInRadians)
	sinAngle := math.Sin(-angleInRadians)

	// rotate point
	rotatedX := translatedX*cosAngle - translatedY*sinAngle
	rotatedY := translatedX*sinAngle + translatedY*cosAngle

	// translate back to original coordinate system
	finalX := rotatedX + rotationCenterX
	finalY := rotatedY + rotationCenterY

	// check against the box's original position (also rotated around the center point)
	// get the box center's position after rotation
	boxOffsetX := boxCenterX - rotationCenterX
	boxOffsetY := boxCenterY - rotationCenterY

	rotatedBoxCenterX := boxOffsetX*math.Cos(angleInRadians) - boxOffsetY*math.Sin(angleInRadians) + rotationCenterX
	rotatedBoxCenterY := boxOffsetX*math.Sin(angleInRadians) + boxOffsetY*math.Cos(angleInRadians) + rotationCenterY

	// check if the point is within the rotated box relative to its new center
	relativeX := finalX - rotatedBoxCenterX
	relativeY := finalY - rotatedBoxCenterY

	// transform relative coordinates to box's coordinate system
	boxX := relativeX*math.Cos(-angleInRadians) - relativeY*math.Sin(-angleInRadians)
	boxY := relativeX*math.Sin(-angleInRadians) + relativeY*math.Cos(-angleInRadians)

	return math.Abs(boxX) <= boxWidth/2 && math.Abs(boxY) <= boxHeight/2
}

func (p *StripesPattern) GetName() string {
	return "stripes"
}

type StripesUpdateRequest struct {
	Parameters StripesParameters `json:"parameters"`
}

func (r *StripesUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *StripesPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &StripesUpdateRequest{
		Parameters: StripesParameters{},
	}
}

func (p *StripesPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
