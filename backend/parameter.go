package main

import (
	"log"
	"math/rand"
)

// Parameter interface defines methods that all parameters must implement
type Parameter interface {
	Randomize()
}

// Add Randomize methods to existing parameter types
func (p *FloatParameter) Randomize() {
	if p.Min != nil {
		oldValue := p.Value
		p.Value = *p.Min + rand.Float64()*(p.Max-*p.Min)
		log.Printf("FloatParameter randomized: min=%v, max=%v, old=%v, new=%v",
			*p.Min, p.Max, oldValue, p.Value)
	}
}

func (p *IntParameter) Randomize() {
	if p.Min != nil {
		oldValue := p.Value
		p.Value = *p.Min + rand.Intn(p.Max-*p.Min+1)
		log.Printf("IntParameter randomized: min=%v, max=%v, old=%v, new=%v",
			*p.Min, p.Max, oldValue, p.Value)
	}
}

func (p *ColorParameter) Randomize() {
	hue := rand.Float64() * 360
	r, g, b := HSVtoRGB(hue, 1.0, 1.0)
	oldColor := p.Value
	p.Value = Color{
		R: colorPigment(int(r * 255)),
		G: colorPigment(int(g * 255)),
		B: colorPigment(int(b * 255)),
	}
	log.Printf("ColorParameter randomized: hue=%v, old=%+v, new=%+v",
		hue, oldColor, p.Value)
}

func (p *BooleanParameter) Randomize() {
	p.Value = rand.Float64() < 0.5
}
