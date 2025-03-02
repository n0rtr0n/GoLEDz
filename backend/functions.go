package main

import (
	"math"
	"math/rand"
)

func buildPixelGrid() *[]Pixel {
	// initialize a grid of pixels for fun coordinate-based pattern development
	pixels := []Pixel{}

	var xPos int16
	var yPos int16
	xStart := 100
	yStart := 100
	spacing := 20
	for i := 0; i < 50; i++ {
		xPos = int16(xStart + i*spacing)
		for j := 0; j < 50; j++ {
			yPos = int16(yStart + j*spacing)
			pixels = append(pixels, Pixel{
				x:               xPos,
				y:               yPos,
				universe:        uint16(i + 1),
				channelPosition: uint16(j + 1),
				color:           Color{R: 0, G: 0, B: 0, W: 0},
			})
		}
	}
	return &pixels
}

func build2ChannelsOfPixels() *[]Pixel {
	pixels := []Pixel{}

	xStart := 100
	yStart := 200
	spacing := 5
	// just two channels for now
	for i := 0; i < 128; i++ {
		xPos := int16(xStart + i*spacing)

		y1Pos := int16(yStart)
		y2Pos := int16(yStart + 20)

		pixels = append(pixels, Pixel{
			x:               xPos,
			y:               y1Pos,
			universe:        1,
			channelPosition: uint16(i + 1),
			color:           Color{R: 0, G: 0, B: 0, W: 0},
		})
		pixels = append(pixels, Pixel{
			x:               xPos,
			y:               y2Pos,
			universe:        3,
			channelPosition: uint16(i + 1),
			color:           Color{R: 0, G: 0, B: 0, W: 0},
		})
	}
	return &pixels
}

func buildMammothSegment(universe uint16, startingChannelPosition uint16, xStart int16, yStart int16, rotationDegrees int16, sections []Section, pixelType PixelType, colorOrder ColorOrder) *[]Pixel {
	pixels := []Pixel{}
	xPos := xStart
	yPos := yStart
	channelPosition := startingChannelPosition
	bigPixelsSpacing := int16(12)
	smallPixelsSpacing := int16(13)
	bigPixelsAlongEachSide := int16(7)
	smallPixelsAlongEachSide := int16(6)

	for i := int16(0); i < bigPixelsAlongEachSide; i++ {
		pixels = append(pixels, Pixel{
			x:               xPos,
			y:               yPos,
			universe:        universe,
			channelPosition: channelPosition,
			sections:        sections,
			pixelType:       pixelType,
			color:           Color{R: 0, G: 0, B: 0, W: 0},
			colorOrder:      colorOrder,
		})
		if i < bigPixelsAlongEachSide-1 {
			xTranslated, yTranslated := rotate(bigPixelsSpacing, 0, rotationDegrees)
			xPos += xTranslated
			yPos += yTranslated
		}
		channelPosition += 1
	}

	// shift these over slightly
	xTranslated, yTranslated := rotate(0, bigPixelsSpacing/2, rotationDegrees)
	xPos += xTranslated
	yPos += yTranslated

	for i := int16(0); i < bigPixelsAlongEachSide; i++ {
		pixels = append(pixels, Pixel{
			x:               xPos,
			y:               yPos,
			universe:        universe,
			channelPosition: channelPosition,
			sections:        sections,
			pixelType:       pixelType,
			color:           Color{R: 0, G: 0, B: 0, W: 0},
			colorOrder:      colorOrder,
		})
		if i < bigPixelsAlongEachSide-1 {
			xTranslated, yTranslated := rotate(-bigPixelsSpacing, 0, rotationDegrees)
			xPos += xTranslated
			yPos += yTranslated
		}
		channelPosition += 1
	}

	// start the snake over but with y += big pixel spacing
	xTranslated, yTranslated = rotate(bigPixelsSpacing/4, bigPixelsSpacing, rotationDegrees)
	xPos = xStart + xTranslated
	yPos = yStart + yTranslated

	for i := int16(0); i < smallPixelsAlongEachSide; i++ {
		pixels = append(pixels, Pixel{
			x:               xPos,
			y:               yPos,
			universe:        universe,
			channelPosition: channelPosition,
			sections:        sections,
			pixelType:       pixelType,
			color:           Color{R: 0, G: 0, B: 0, W: 0},
			colorOrder:      colorOrder,
		})
		xTranslated, yTranslated := rotate(smallPixelsSpacing, 0, rotationDegrees)
		xPos += xTranslated
		yPos += yTranslated
		channelPosition += 1
	}

	return &pixels
}

func buildTuskSegment(universe uint16, startingChannelNumber uint16, xStart int16, yStart int16, rotationDegrees int16, sections []Section, pixelType PixelType, colorOrder ColorOrder) *[]Pixel {
	pixels := []Pixel{}
	xPos := xStart
	yPos := yStart
	channelPosition := startingChannelNumber
	pixelsSpacing := int16(5)
	totalPixels := int16(60)

	for i := int16(0); i < totalPixels; i++ {
		pixels = append(pixels, Pixel{
			x:               xPos,
			y:               yPos,
			universe:        universe,
			channelPosition: channelPosition,
			sections:        sections,
			pixelType:       pixelType,
			color:           Color{R: 0, G: 0, B: 0, W: 0},
			colorOrder:      colorOrder,
		})
		xTranslated, yTranslated := rotate(pixelsSpacing, 0, rotationDegrees)
		xPos += xTranslated
		yPos += yTranslated
		channelPosition += 1
	}
	return &pixels
}

func rotate(x int16, y int16, rotationDegrees int16) (int16, int16) {
	radians := degreesToRadians(float64(rotationDegrees))
	newX := int16(float64(x)*math.Cos(float64(radians))) + int16(float64(y)*math.Sin(float64(radians)))
	newY := int16(float64(y)*math.Cos(float64(radians))) + int16(float64(-x)*math.Sin(float64(radians)))

	return newX, newY
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// chance of returning true
func randomChancePercent(chance float64) bool {
	return (rand.Float64() * 100.0) <= chance
}

// helper function to return address of float value
// this is needed in order to allow values of 0.0, while specifying "omitempty"
// in JSON marshalling
func floatPointer(value float64) *float64 {
	return &value
}

// same as above for int pointer. effectively allows us to set
// a minimum value of 0 for params
func intPointer(value int) *int {
	return &value
}
