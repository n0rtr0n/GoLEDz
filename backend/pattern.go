package main

type Pattern interface {
	Update()
}

type SolidColorPattern struct {
	pixelMap *PixelMap
	color    Color
}

func (p *SolidColorPattern) Update() {
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = p.color
	}
}
