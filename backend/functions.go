package main

func buildPixelGrid() *[]Pixel {
	// initialize a grid of pixels for fun coordinate-based pattern development
	pixels := []Pixel{}

	var xPos int16
	var yPos int16
	xStart := 100
	yStart := 100
	spacing := 20
	for i := 0; i < 20; i++ {
		xPos = int16(xStart + i*spacing)
		for j := 0; j < 20; j++ {
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
