package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type RipplePattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters RippleParameters `json:"parameters"`
	Label      string           `json:"label,omitempty"`
	ripples    []ripple
	lastUpdate time.Time
}

type ripple struct {
	center    Point
	radius    float64
	maxRadius float64
	strength  float64 // initial strength of the ripple
	age       float64 // age of the ripple in seconds
}

func (p *RipplePattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RippleParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.RippleCount.Update(newParams.RippleCount.Value)
	p.Parameters.RippleWidth.Update(newParams.RippleWidth.Value)
	p.Parameters.RippleLifetime.Update(newParams.RippleLifetime.Value)
	p.Parameters.BackgroundColor.Update(newParams.BackgroundColor.Value)
	p.Parameters.AutoGenerate.Update(newParams.AutoGenerate.Value)
	return nil
}

type RippleParameters struct {
	Speed           FloatParameter   `json:"speed"`
	RippleCount     IntParameter     `json:"rippleCount"`
	RippleWidth     FloatParameter   `json:"rippleWidth"`
	RippleLifetime  FloatParameter   `json:"rippleLifetime"`
	BackgroundColor ColorParameter   `json:"backgroundColor"`
	AutoGenerate    BooleanParameter `json:"autoGenerate"`
}

func (p *RipplePattern) Update() {
	// initialize if this is the first update
	if p.lastUpdate.IsZero() {
		p.lastUpdate = time.Now()
		p.ripples = make([]ripple, 0)

		// create initial ripples
		for i := 0; i < p.Parameters.RippleCount.Value; i++ {
			p.addRandomRipple()
		}
	}

	// calculate time delta
	now := time.Now()
	deltaTime := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now

	// update existing ripples
	speed := p.Parameters.Speed.Value
	rippleWidth := p.Parameters.RippleWidth.Value
	backgroundColor := p.Parameters.BackgroundColor.Value
	maxLifetime := p.Parameters.RippleLifetime.Value

	// start with background color
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = backgroundColor
	}

	// keep track of active ripples
	activeRipples := make([]ripple, 0, len(p.ripples))

	// update and draw each ripple
	for _, r := range p.ripples {
		// update ripple
		r.radius += speed * deltaTime * 50.0 // scale speed for better control
		r.age += deltaTime

		// skip expired ripples
		if r.age >= maxLifetime || r.radius >= r.maxRadius {
			continue
		}

		// calculate ripple strength based on age (fade out as it gets older)
		lifeProgress := r.age / maxLifetime
		fadeStrength := 1.0 - lifeProgress

		activeRipples = append(activeRipples, r)

		// draw the ripple
		for i, pixel := range *p.pixelMap.pixels {
			point := Point{pixel.x, pixel.y}
			dist := distanceBetweenPoints(point, r.center)

			// calculate ripple effect - creates a ring shape
			ringEffect := math.Exp(-math.Pow(dist-r.radius, 2) / (2 * rippleWidth * rippleWidth))
			ringEffect *= fadeStrength * r.strength

			// only apply effect if it's significant
			if ringEffect > 0.05 {
				// get color from mask if available
				if p.GetColorMask() != nil {
					maskColor := p.GetColorMask().GetColorAt(point)

					// blend with background based on ring effect
					blendedColor := blendColors(backgroundColor, maskColor, ringEffect)
					(*p.pixelMap.pixels)[i].color = blendedColor
				}
			}
		}
	}

	p.ripples = activeRipples

	// auto-generate new ripples if needed
	if p.Parameters.AutoGenerate.Value && len(p.ripples) < p.Parameters.RippleCount.Value {
		// add new ripples to maintain the desired count
		for i := 0; i < p.Parameters.RippleCount.Value-len(p.ripples); i++ {
			p.addRandomRipple()
		}
	}

	// update the color mask if we have one
	if p.GetColorMask() != nil {
		p.GetColorMask().Update()
	}
}

func (p *RipplePattern) addRandomRipple() {
	// create a ripple at a random position
	center := Point{
		X: int16(rand.Intn(MAX_X)),
		Y: int16(rand.Intn(MAX_Y)),
	}

	// calculate maximum radius based on distance to furthest corner
	corners := []Point{
		{0, 0},
		{int16(MAX_X), 0},
		{0, int16(MAX_Y)},
		{int16(MAX_X), int16(MAX_Y)},
	}

	maxRadius := 0.0
	for _, corner := range corners {
		dist := distanceBetweenPoints(center, corner)
		if dist > maxRadius {
			maxRadius = dist
		}
	}

	// add the new ripple
	p.ripples = append(p.ripples, ripple{
		center:    center,
		radius:    0,
		maxRadius: maxRadius,
		strength:  0.5 + rand.Float64()*0.5,
		age:       0,
	})
}

func (p *RipplePattern) GetName() string {
	return "ripple"
}

type RippleUpdateRequest struct {
	Parameters RippleParameters `json:"parameters"`
}

func (r *RippleUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *RipplePattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RippleUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *RipplePattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
