package main

import (
	"fmt"
	"math"
)

func buildPixelGrid() *[]Pixel {
	// initialize a grid of pixels for fun coordinate-based pattern development
	pixels := []Pixel{}

	var xPos int16
	var yPos int16
	xStart := 100
	yStart := 100
	spacing := 10
	for i := 0; i < 50; i++ {
		xPos = int16(xStart + i*spacing)
		for j := 0; j < 50; j++ {
			yPos = int16(yStart + j*spacing)
			pixels = append(pixels, Pixel{x: xPos, y: yPos, universe: uint16(i + 1), channelPosition: uint16(j + 1)})
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
	for i := 0; i < 150; i++ {
		xPos := int16(xStart + i*spacing)

		y1Pos := int16(yStart)
		y2Pos := int16(yStart + 20)

		pixels = append(pixels, Pixel{x: xPos, y: y1Pos, universe: 1, channelPosition: uint16(i + 1)})
		pixels = append(pixels, Pixel{x: xPos, y: y2Pos, universe: 3, channelPosition: uint16(i + 1)})
	}
	return &pixels
}

func buildLegSegment(startingChannelNumber uint16, universe uint16, xStart int16, yStart int16, rotationDegrees int16) *[]Pixel {

	pixels := []Pixel{}
	xPos := xStart
	yPos := yStart
	channelPosition := startingChannelNumber
	bigPixelsSpacing := int16(40)
	smallPixelsSpacing := int16(12)
	bigPixelsAlongEachSide := int16(4)
	smallPixelsAlongEachSide := int16(12)

	for i := int16(0); i < bigPixelsAlongEachSide; i++ {
		xPos = xStart + i*bigPixelsSpacing

		xTranslated, yTranslated := translate(xPos, yPos, rotationDegrees)
		pixels = append(pixels, Pixel{x: xTranslated, y: yTranslated, universe: universe, channelPosition: channelPosition})
		channelPosition += 1
	}
	yPos += bigPixelsSpacing
	for i := int16(0); i < bigPixelsAlongEachSide; i++ {
		xPos = xStart + (bigPixelsAlongEachSide-i-1)*bigPixelsSpacing
		xTranslated, yTranslated := translate(xPos, yPos, rotationDegrees)
		pixels = append(pixels, Pixel{x: xTranslated, y: yTranslated, universe: universe, channelPosition: channelPosition})
		channelPosition += 1
	}
	yPos += bigPixelsSpacing

	for i := int16(0); i < smallPixelsAlongEachSide; i++ {
		xPos = xStart + i*smallPixelsSpacing
		xTranslated, yTranslated := translate(xPos, yPos, rotationDegrees)
		pixels = append(pixels, Pixel{x: xTranslated, y: yTranslated, universe: universe, channelPosition: channelPosition})
		channelPosition += 1
	}
	fmt.Println(pixels)

	return &pixels
}

func translate(x int16, y int16, rotationDegrees int16) (newX int16, newY int16) {
	newX = int16(float64(-y)*math.Sin(float64(rotationDegrees))) + int16(float64(x)*math.Cos(float64(rotationDegrees)))
	newY = int16(float64(-x)*math.Sin(float64(rotationDegrees))) + int16(float64(y)*math.Cos(float64(rotationDegrees)))
	fmt.Println(x, y)
	fmt.Println("changes to")
	fmt.Println(newX, newY)
	return newX, newY
}
