package main

import "encoding/json"

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
