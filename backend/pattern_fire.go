package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type FirePattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters FireParameters `json:"parameters"`
	Label      string         `json:"label,omitempty"`
	heatMap    []float64
	lastUpdate time.Time
}

type FireParameters struct {
	Cooling       FloatParameter `json:"cooling"`
	Sparking      FloatParameter `json:"sparking"`
	Speed         FloatParameter `json:"speed"`
	ColorScheme   IntParameter   `json:"colorScheme"`
	WindDirection FloatParameter `json:"windDirection"`
	WindStrength  FloatParameter `json:"windStrength"`
}

func (p *FirePattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(FireParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Cooling.Update(newParams.Cooling.Value)
	p.Parameters.Sparking.Update(newParams.Sparking.Value)
	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.ColorScheme.Update(newParams.ColorScheme.Value)
	p.Parameters.WindDirection.Update(newParams.WindDirection.Value)
	p.Parameters.WindStrength.Update(newParams.WindStrength.Value)
	return nil
}

func (p *FirePattern) Update() {
	// Initialize if this is the first update
	if p.lastUpdate.IsZero() {
		p.lastUpdate = time.Now()
		p.initializeHeatMap()
	}

	// Calculate time delta
	now := time.Now()
	deltaTime := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now

	// Get parameters
	cooling := p.Parameters.Cooling.Value
	sparking := p.Parameters.Sparking.Value
	speed := p.Parameters.Speed.Value
	colorScheme := p.Parameters.ColorScheme.Value
	windDirection := p.Parameters.WindDirection.Value
	windStrength := p.Parameters.WindStrength.Value

	// Adjust for speed
	iterations := int(speed * deltaTime * 10)
	if iterations < 1 {
		iterations = 1
	}

	// Run the fire simulation multiple times based on speed
	for i := 0; i < iterations; i++ {
		p.simulateFire(cooling, sparking, windDirection, windStrength)
	}

	// Map heat to colors and update pixels
	p.mapHeatToColors(colorScheme)

	// Update the color mask if we have one
	if p.GetColorMask() != nil {
		p.GetColorMask().Update()
	}
}

func (p *FirePattern) initializeHeatMap() {
	// Create a heat map for each pixel
	p.heatMap = make([]float64, len(*p.pixelMap.pixels))

	// Find the maximum Y value
	maxY := int16(0)
	for _, pixel := range *p.pixelMap.pixels {
		if pixel.y > maxY {
			maxY = pixel.y
		}
	}

	// Initialize ALL pixels with substantial heat
	// More heat at the bottom, gradually decreasing toward the top
	for i, pixel := range *p.pixelMap.pixels {
		// Calculate relative height (0 at top, 1 at bottom)
		relativeHeight := float64(pixel.y) / float64(maxY)

		// Use a curve that gives more heat to all pixels
		// Even the top pixels get at least 30% heat
		baseHeat := 0.3 + 0.7*relativeHeight

		// Add some randomness
		randomFactor := 0.7 + 0.3*rand.Float64()

		// Set the heat value
		p.heatMap[i] = baseHeat * randomFactor
	}
}

func (p *FirePattern) simulateFire(cooling, sparking, windDirection, windStrength float64) {
	// Reduce cooling to keep more heat throughout the display
	adjustedCooling := cooling * 0.7

	// Cool down every cell a little
	for i := range p.heatMap {
		cooldown := rand.Float64() * adjustedCooling * 0.1
		if p.heatMap[i] > cooldown {
			p.heatMap[i] -= cooldown
		} else {
			p.heatMap[i] = 0
		}
	}

	// Create a mapping from pixel coordinates to heat map indices
	pixelToHeatIndex := make(map[Point]int)
	for i, pixel := range *p.pixelMap.pixels {
		pixelToHeatIndex[Point{pixel.x, pixel.y}] = i
	}

	// Heat rises - for each pixel, find pixels above it and transfer heat
	for i, pixel := range *p.pixelMap.pixels {
		if p.heatMap[i] > 0 {
			// Find pixels above this one (lower y value)
			aboveY := pixel.y - 1

			// Apply wind by shifting the x coordinate
			windOffset := int16(math.Sin(windDirection*math.Pi/180) * windStrength)
			aboveX := pixel.x + windOffset

			// Find the pixel at this position
			abovePoint := Point{aboveX, aboveY}
			if aboveIndex, exists := pixelToHeatIndex[abovePoint]; exists {
				// Transfer more heat upward (50% instead of 40%)
				heatTransfer := p.heatMap[i] * 0.5
				p.heatMap[i] -= heatTransfer
				p.heatMap[aboveIndex] += heatTransfer
			}
		}
	}

	// Randomly ignite new sparks throughout the ENTIRE display
	if rand.Float64() < sparking {
		// Ignite multiple random pixels across the entire display
		sparkCount := 8 + rand.Intn(8) // 8-15 sparks per iteration
		for s := 0; s < sparkCount; s++ {
			// Pick a random pixel
			idx := rand.Intn(len(p.heatMap))
			pixel := (*p.pixelMap.pixels)[idx]

			// Higher heat for pixels near the bottom
			maxY := int16(0)
			for _, p := range *p.pixelMap.pixels {
				if p.y > maxY {
					maxY = p.y
				}
			}

			relativeHeight := float64(pixel.y) / float64(maxY)

			// Base heat value depends on height
			baseHeat := 0.5 + 0.5*relativeHeight

			// Add randomness
			heatValue := baseHeat + rand.Float64()*0.3

			// Set the heat
			p.heatMap[idx] = math.Min(1.0, heatValue)
		}
	}
}

func (p *FirePattern) mapHeatToColors(colorScheme int) {
	for i, heat := range p.heatMap {
		var color Color

		// Map heat value to color based on color scheme
		switch colorScheme {
		case 0: // Classic fire (red-yellow)
			color = p.heatToFireColor(heat)
		case 1: // Blue fire
			color = p.heatToBlueFireColor(heat)
		case 2: // Green fire
			color = p.heatToGreenFireColor(heat)
		case 3: // Purple fire
			color = p.heatToPurpleFireColor(heat)
		default:
			color = p.heatToFireColor(heat)
		}

		// Apply color mask if available
		if p.GetColorMask() != nil {
			pixel := (*p.pixelMap.pixels)[i]
			maskColor := p.GetColorMask().GetColorAt(Point{pixel.x, pixel.y})

			// Blend with mask color based on heat
			color = Color{
				R: colorPigment(float64(color.R)*0.7 + float64(maskColor.R)*0.3*heat),
				G: colorPigment(float64(color.G)*0.7 + float64(maskColor.G)*0.3*heat),
				B: colorPigment(float64(color.B)*0.7 + float64(maskColor.B)*0.3*heat),
				W: 0,
			}
		}

		(*p.pixelMap.pixels)[i].color = color
	}
}

func (p *FirePattern) heatToFireColor(heat float64) Color {
	// Classic fire colors: black -> red -> orange -> yellow -> white
	heat = math.Min(1.0, math.Max(0.0, heat))

	if heat < 0.25 {
		// Black to red
		intensity := heat * 4
		return Color{
			R: colorPigment(255 * intensity),
			G: 0,
			B: 0,
			W: 0,
		}
	} else if heat < 0.5 {
		// Red to orange
		intensity := (heat - 0.25) * 4
		return Color{
			R: 255,
			G: colorPigment(165 * intensity),
			B: 0,
			W: 0,
		}
	} else if heat < 0.75 {
		// Orange to yellow
		intensity := (heat - 0.5) * 4
		return Color{
			R: 255,
			G: colorPigment(165 + (255-165)*intensity),
			B: 0,
			W: 0,
		}
	} else {
		// Yellow to white
		intensity := (heat - 0.75) * 4
		return Color{
			R: 255,
			G: 255,
			B: colorPigment(255 * intensity),
			W: 0,
		}
	}
}

func (p *FirePattern) heatToBlueFireColor(heat float64) Color {
	// Blue fire colors: black -> deep blue -> blue -> light blue -> white
	heat = math.Min(1.0, math.Max(0.0, heat))

	if heat < 0.25 {
		// Black to deep blue
		intensity := heat * 4
		return Color{
			R: 0,
			G: 0,
			B: colorPigment(128 * intensity),
			W: 0,
		}
	} else if heat < 0.5 {
		// Deep blue to blue
		intensity := (heat - 0.25) * 4
		return Color{
			R: 0,
			G: colorPigment(64 * intensity),
			B: colorPigment(128 + (255-128)*intensity),
			W: 0,
		}
	} else if heat < 0.75 {
		// Blue to light blue
		intensity := (heat - 0.5) * 4
		return Color{
			R: colorPigment(64 * intensity),
			G: colorPigment(64 + (192-64)*intensity),
			B: 255,
			W: 0,
		}
	} else {
		// Light blue to white
		intensity := (heat - 0.75) * 4
		return Color{
			R: colorPigment(64 + (255-64)*intensity),
			G: colorPigment(192 + (255-192)*intensity),
			B: 255,
			W: 0,
		}
	}
}

func (p *FirePattern) heatToGreenFireColor(heat float64) Color {
	// Green fire colors
	heat = math.Min(1.0, math.Max(0.0, heat))

	if heat < 0.25 {
		// Black to dark green
		intensity := heat * 4
		return Color{
			R: 0,
			G: colorPigment(100 * intensity),
			B: 0,
			W: 0,
		}
	} else if heat < 0.5 {
		// Dark green to green
		intensity := (heat - 0.25) * 4
		return Color{
			R: 0,
			G: colorPigment(100 + (200-100)*intensity),
			B: 0,
			W: 0,
		}
	} else if heat < 0.75 {
		// Green to yellow-green
		intensity := (heat - 0.5) * 4
		return Color{
			R: colorPigment(180 * intensity),
			G: colorPigment(200 + (255-200)*intensity),
			B: 0,
			W: 0,
		}
	} else {
		// Yellow-green to white
		intensity := (heat - 0.75) * 4
		return Color{
			R: colorPigment(180 + (255-180)*intensity),
			G: 255,
			B: colorPigment(220 * intensity),
			W: 0,
		}
	}
}

func (p *FirePattern) heatToPurpleFireColor(heat float64) Color {
	// Purple fire colors
	heat = math.Min(1.0, math.Max(0.0, heat))

	if heat < 0.25 {
		// Black to dark purple
		intensity := heat * 4
		return Color{
			R: colorPigment(80 * intensity),
			G: 0,
			B: colorPigment(100 * intensity),
			W: 0,
		}
	} else if heat < 0.5 {
		// Dark purple to purple
		intensity := (heat - 0.25) * 4
		return Color{
			R: colorPigment(80 + (150-80)*intensity),
			G: 0,
			B: colorPigment(100 + (200-100)*intensity),
			W: 0,
		}
	} else if heat < 0.75 {
		// Purple to pink
		intensity := (heat - 0.5) * 4
		return Color{
			R: colorPigment(150 + (255-150)*intensity),
			G: colorPigment(100 * intensity),
			B: colorPigment(200 + (255-200)*intensity),
			W: 0,
		}
	} else {
		// Pink to white
		intensity := (heat - 0.75) * 4
		return Color{
			R: 255,
			G: colorPigment(100 + (255-100)*intensity),
			B: 255,
			W: 0,
		}
	}
}

func (p *FirePattern) GetName() string {
	return "fire"
}

type FireUpdateRequest struct {
	Parameters FireParameters `json:"parameters"`
}

func (r *FireUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *FirePattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &FireUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *FirePattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
