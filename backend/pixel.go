package main

import (
	"encoding/json"
	"math"
)

type ColorOrder uint8

// pixels from different manufacturers often come with surprises
// surprise! the color ordering is sometimes different, and we have to account for that in code
const (
	RGB ColorOrder = iota
	RBG
	BRG
	BGR
	GRB
	GBR
)

const CENTER_X = 400
const CENTER_Y = 400

type Pixel struct {
	x               int16
	y               int16
	color           Color
	colorOrder      ColorOrder
	universe        uint16
	channelPosition uint16
}

type PixelMap struct {
	pixels     *[]Pixel
	brightness uint8
}

type Point struct {
	X, Y int16
}

func calculateAngle(target Point, center Point) float64 {
	// difference from center point
	dx := target.X - center.X
	dy := center.Y - target.Y

	// calculate angle in radians using atan2
	angleRadians := math.Atan2(float64(dy), float64(dx))

	// convert to degrees
	angleDegrees := angleRadians * 180 / math.Pi

	// normalize to 0-360 range
	if angleDegrees < 0 {
		angleDegrees += 360
	}

	return angleDegrees
}

func (p *PixelMap) toJSON() ([]byte, error) {

	var data []map[string]interface{}

	for _, pixel := range *p.pixels {
		newPixel := map[string]interface{}{
			"x": pixel.x,
			"y": pixel.y,
			"r": pixel.color.R,
			"g": pixel.color.G,
			"b": pixel.color.B,
		}
		data = append(data, newPixel)
	}

	pixelData := map[string]interface{}{
		"pixels": data,
	}

	return json.Marshal(pixelData)
}
