package main

import "encoding/json"

// RBG values 0 to 255
type Color struct {
	r uint8
	g uint8
	b uint8
}

func (c *Color) toString() []byte {
	return []byte{
		byte(c.r),
		byte(c.g),
		byte(c.b),
	}
}

type Pixel struct {
	x               int16
	y               int16
	color           Color
	universe        uint16
	channelPosition uint16
}

type PixelMap struct {
	pixels *[]Pixel
}

func (p *PixelMap) toJSON() ([]byte, error) {

	var data []map[string]interface{}

	for _, pixel := range *p.pixels {
		newPixel := map[string]interface{}{
			"x": pixel.x,
			"y": pixel.y,
			"r": pixel.color.r,
			"g": pixel.color.g,
			"b": pixel.color.b,
		}
		data = append(data, newPixel)
	}

	pixelData := map[string]interface{}{
		"pixels": data,
	}

	return json.Marshal(pixelData)
}
