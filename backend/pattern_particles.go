package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type ParticlePattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters ParticleParameters `json:"parameters"`
	Label      string             `json:"label,omitempty"`
	particles  []particle
	lastUpdate time.Time
}

type particle struct {
	x, y         float64
	vx, vy       float64
	age          float64
	lifetime     float64
	size         float64
	color        Color
	initialColor Color
	finalColor   Color
}

type ParticleParameters struct {
	EmissionRate  FloatParameter `json:"emissionRate"`
	ParticleLife  FloatParameter `json:"particleLife"`
	Gravity       FloatParameter `json:"gravity"`
	InitialColor  ColorParameter `json:"initialColor"`
	FinalColor    ColorParameter `json:"finalColor"`
	ParticleSize  FloatParameter `json:"particleSize"`
	EmitterX      FloatParameter `json:"emitterX"`    // 0-1 range
	EmitterY      FloatParameter `json:"emitterY"`    // 0-1 range
	SpreadAngle   FloatParameter `json:"spreadAngle"` // degrees
	ParticleSpeed FloatParameter `json:"particleSpeed"`
}

func (p *ParticlePattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(ParticleParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.EmissionRate.Update(newParams.EmissionRate.Value)
	p.Parameters.ParticleLife.Update(newParams.ParticleLife.Value)
	p.Parameters.Gravity.Update(newParams.Gravity.Value)
	p.Parameters.InitialColor.Update(newParams.InitialColor.Value)
	p.Parameters.FinalColor.Update(newParams.FinalColor.Value)
	p.Parameters.ParticleSize.Update(newParams.ParticleSize.Value)
	p.Parameters.EmitterX.Update(newParams.EmitterX.Value)
	p.Parameters.EmitterY.Update(newParams.EmitterY.Value)
	p.Parameters.SpreadAngle.Update(newParams.SpreadAngle.Value)
	p.Parameters.ParticleSpeed.Update(newParams.ParticleSpeed.Value)
	return nil
}

func (p *ParticlePattern) Update() {
	// initialize if this is the first update
	if p.lastUpdate.IsZero() {
		p.lastUpdate = time.Now()
		p.particles = make([]particle, 0)
	}

	// calculate time delta
	now := time.Now()
	deltaTime := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now

	// get parameters
	emissionRate := p.Parameters.EmissionRate.Value
	particleLife := p.Parameters.ParticleLife.Value
	gravity := p.Parameters.Gravity.Value
	initialColor := p.Parameters.InitialColor.Value
	finalColor := p.Parameters.FinalColor.Value
	particleSize := p.Parameters.ParticleSize.Value
	emitterX := p.Parameters.EmitterX.Value
	emitterY := p.Parameters.EmitterY.Value
	spreadAngle := p.Parameters.SpreadAngle.Value
	particleSpeed := p.Parameters.ParticleSpeed.Value

	// find max X and Y values
	maxX, maxY := int16(0), int16(0)
	for _, pixel := range *p.pixelMap.pixels {
		if pixel.x > maxX {
			maxX = pixel.x
		}
		if pixel.y > maxY {
			maxY = pixel.y
		}
	}

	// calculate emitter position in pixel coordinates
	emitterPosX := float64(maxX) * emitterX
	emitterPosY := float64(maxY) * emitterY

	// clear all pixels to black
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = Color{R: 0, G: 0, B: 0, W: 0}
	}

	// emit new particles
	particlesToEmit := int(emissionRate * deltaTime)
	for i := 0; i < particlesToEmit; i++ {
		// calculate random angle within spread
		angle := (rand.Float64()*spreadAngle - spreadAngle/2) * math.Pi / 180

		// calculate velocity components
		speed := particleSpeed * (0.8 + rand.Float64()*0.4) // vary speed slightly
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed

		// create new particle
		p.particles = append(p.particles, particle{
			x:            emitterPosX,
			y:            emitterPosY,
			vx:           vx,
			vy:           vy,
			age:          0,
			lifetime:     particleLife * (0.8 + rand.Float64()*0.4), // vary lifetime slightly
			size:         particleSize * (0.8 + rand.Float64()*0.4), // vary size slightly
			initialColor: initialColor,
			finalColor:   finalColor,
		})
	}

	// update existing particles
	var activeParticles []particle
	for _, part := range p.particles {
		// update position
		part.x += part.vx * deltaTime
		part.y += part.vy * deltaTime

		// apply gravity
		part.vy += gravity * deltaTime

		// update age
		part.age += deltaTime

		// calculate color based on age
		lifeProgress := part.age / part.lifetime
		if lifeProgress < 1.0 {
			// blend from initial to final color
			part.color = Color{
				R: colorPigment(float64(part.initialColor.R)*(1-lifeProgress) + float64(part.finalColor.R)*lifeProgress),
				G: colorPigment(float64(part.initialColor.G)*(1-lifeProgress) + float64(part.finalColor.G)*lifeProgress),
				B: colorPigment(float64(part.initialColor.B)*(1-lifeProgress) + float64(part.finalColor.B)*lifeProgress),
				W: 0,
			}

			// keep particle if still alive
			activeParticles = append(activeParticles, part)
		}
	}
	p.particles = activeParticles

	// draw all particles
	for _, part := range p.particles {
		radius := part.size

		for i, pixel := range *p.pixelMap.pixels {
			// calculate distance from pixel to particle center
			dx := float64(pixel.x) - part.x
			dy := float64(pixel.y) - part.y
			distance := math.Sqrt(dx*dx + dy*dy)

			// skip pixels outside particle's influence
			if distance > radius {
				continue
			}

			// calculate intensity based on distance
			intensity := 1.0 - distance/radius
			intensity = math.Pow(intensity, 0.5) // adjust power for softer/harder edge

			// apply color mask if available
			if p.GetColorMask() != nil {
				maskColor := p.GetColorMask().GetColorAt(Point{pixel.x, pixel.y})

				// blend with mask color
				color := Color{
					R: colorPigment(float64(part.color.R)*intensity + float64(maskColor.R)*(1-intensity)),
					G: colorPigment(float64(part.color.G)*intensity + float64(maskColor.G)*(1-intensity)),
					B: colorPigment(float64(part.color.B)*intensity + float64(maskColor.B)*(1-intensity)),
					W: 0,
				}

				// blend with existing color (additive blending)
				existingColor := (*p.pixelMap.pixels)[i].color
				(*p.pixelMap.pixels)[i].color = Color{
					R: colorPigment(math.Min(255, float64(existingColor.R)+float64(color.R))),
					G: colorPigment(math.Min(255, float64(existingColor.G)+float64(color.G))),
					B: colorPigment(math.Min(255, float64(existingColor.B)+float64(color.B))),
					W: 0,
				}
			} else {
				// blend with existing color (additive blending)
				existingColor := (*p.pixelMap.pixels)[i].color
				(*p.pixelMap.pixels)[i].color = Color{
					R: colorPigment(math.Min(255, float64(existingColor.R)+float64(part.color.R)*intensity)),
					G: colorPigment(math.Min(255, float64(existingColor.G)+float64(part.color.G)*intensity)),
					B: colorPigment(math.Min(255, float64(existingColor.B)+float64(part.color.B)*intensity)),
					W: 0,
				}
			}
		}
	}

	// update the color mask if we have one
	if p.GetColorMask() != nil {
		p.GetColorMask().Update()
	}
}

func (p *ParticlePattern) GetName() string {
	return "particles"
}

type ParticleUpdateRequest struct {
	Parameters ParticleParameters `json:"parameters"`
}

func (r *ParticleUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *ParticlePattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &ParticleUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *ParticlePattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
