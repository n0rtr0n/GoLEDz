package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

const MAX_DROPS = 5000

type MatrixPattern struct {
	BasePattern
	pixelMap   *PixelMap
	Parameters MatrixParameters `json:"parameters"`
	Label      string           `json:"label,omitempty"`
	drops      []matrixDrop
	lastUpdate time.Time
}

type matrixDrop struct {
	x        int16
	length   int
	speed    float64
	position float64
	bright   float64
}

type MatrixParameters struct {
	Speed      FloatParameter   `json:"speed"`
	Density    FloatParameter   `json:"density"`
	DropLength FloatParameter   `json:"dropLength"`
	Reversed   BooleanParameter `json:"reversed"`
}

func (p *MatrixPattern) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(MatrixParameters)
	if !ok {
		err := fmt.Sprintf("Could not cast updated parameters for %v pattern", p.GetName())
		return errors.New(err)
	}

	p.Parameters.Speed.Update(newParams.Speed.Value)
	p.Parameters.Density.Update(newParams.Density.Value)
	p.Parameters.DropLength.Update(newParams.DropLength.Value)
	p.Parameters.Reversed.Update(newParams.Reversed.Value)
	return nil
}

func (p *MatrixPattern) Update() {
	// nitialize if this is the first update
	if p.lastUpdate.IsZero() {
		p.lastUpdate = time.Now()
		p.drops = make([]matrixDrop, 0)
	}

	// calculate time delta
	now := time.Now()
	deltaTime := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now

	// get parameters
	speed := p.Parameters.Speed.Value
	density := p.Parameters.Density.Value
	dropLength := p.Parameters.DropLength.Value
	reversed := p.Parameters.Reversed.Value

	// build pixel lookup and find boundaries
	pixelLookup, minX, minY, maxX, maxY := p.buildPixelLookup()

	// clear all pixels to black
	p.clearPixels()

	// get all unique X coordinates
	xCoords := p.getUniqueXCoordinates(pixelLookup)

	// reset drops if there are too many to prevent glitches
	if len(p.drops) > len(xCoords)*2 || len(p.drops) > MAX_DROPS {
		p.drops = make([]matrixDrop, 0)
	}

	// update existing drops
	p.updateExistingDrops(deltaTime, speed, reversed, minY, maxY)

	// add new drops as needed
	p.addNewDrops(xCoords, density, dropLength, deltaTime, reversed, minX, minY, maxX, maxY)

	// draw all drops
	p.drawDrops(pixelLookup, reversed)

	// add random sparkles
	p.addSparkles(density)

	// update the color mask if we have one
	if p.GetColorMask() != nil {
		p.GetColorMask().Update()
	}
}

// Helper functions to reduce complexity and duplication

func (p *MatrixPattern) buildPixelLookup() (map[int16]map[int16]int, int16, int16, int16, int16) {
	maxX, maxY := int16(0), int16(0)
	minX, minY := int16(9999), int16(9999)
	pixelLookup := make(map[int16]map[int16]int)

	for i, pixel := range *p.pixelMap.pixels {
		// track min/max coordinates
		if pixel.x > maxX {
			maxX = pixel.x
		}
		if pixel.y > maxY {
			maxY = pixel.y
		}
		if pixel.x < minX {
			minX = pixel.x
		}
		if pixel.y < minY {
			minY = pixel.y
		}

		// build lookup map for faster pixel access
		if _, exists := pixelLookup[pixel.x]; !exists {
			pixelLookup[pixel.x] = make(map[int16]int)
		}
		pixelLookup[pixel.x][pixel.y] = i
	}

	return pixelLookup, minX, minY, maxX, maxY
}

func (p *MatrixPattern) clearPixels() {
	for i := range *p.pixelMap.pixels {
		(*p.pixelMap.pixels)[i].color = Color{R: 0, G: 0, B: 0, W: 0}
	}
}

func (p *MatrixPattern) getUniqueXCoordinates(pixelLookup map[int16]map[int16]int) []int16 {
	xCoords := make([]int16, 0, len(pixelLookup))
	for x := range pixelLookup {
		xCoords = append(xCoords, x)
	}

	// sort for even distribution
	sort.Slice(xCoords, func(i, j int) bool {
		return xCoords[i] < xCoords[j]
	})

	return xCoords
}

func (p *MatrixPattern) updateExistingDrops(deltaTime, speed float64, reversed bool, minY, maxY int16) {
	var activeDrops []matrixDrop
	for _, drop := range p.drops {
		// update position based on direction
		if reversed {
			drop.position -= drop.speed * speed * deltaTime
		} else {
			drop.position += drop.speed * speed * deltaTime
		}

		if reversed {
			// for upward movement
			if drop.position+float64(drop.length) > 0 {
				activeDrops = append(activeDrops, drop)
			} else {
				// loop back to bottom
				drop.position = float64(maxY) + float64(drop.length)
				drop.bright = 0.8 + rand.Float64()*0.2
				activeDrops = append(activeDrops, drop)
			}
		} else {
			// for downward movement
			if drop.position-float64(drop.length) < float64(maxY) {
				activeDrops = append(activeDrops, drop)
			} else {
				// loop back to top
				drop.position = float64(minY)
				drop.bright = 0.8 + rand.Float64()*0.2
				activeDrops = append(activeDrops, drop)
			}
		}
	}
	p.drops = activeDrops
}

func (p *MatrixPattern) addNewDrops(xCoords []int16, density, dropLength, deltaTime float64, reversed bool, minX, minY, maxX, maxY int16) {
	// ensure enough drops to cover the canvas
	xRange := int(maxX - minX)
	desiredDropCount := xRange / 10 // one drop every ~10 pixels

	// track which X coordinates already have drops
	existingDropX := make(map[int16]bool)
	for _, drop := range p.drops {
		existingDropX[drop.x] = true
	}

	// add drops at X coordinates that don't have drops yet
	if len(p.drops) < desiredDropCount {
		for _, x := range xCoords {
			if !existingDropX[x] && rand.Float64() < 0.5 {
				length := int(10 + rand.Float64()*dropLength)
				startY := minY + int16(rand.Float64()*float64(maxY-minY))

				initialPosition := float64(0)
				if reversed {
					initialPosition = float64(maxY)
				} else {
					initialPosition = float64(startY)
				}

				p.drops = append(p.drops, matrixDrop{
					x:        x,
					length:   length,
					speed:    20 + rand.Float64()*30,
					position: initialPosition,
					bright:   0.8 + rand.Float64()*0.2,
				})

				existingDropX[x] = true
			}
		}
	}

	// add new random drops based on density
	if len(p.drops) < len(xCoords)*2 {
		dropChance := density * deltaTime * 5

		for i := 0; i < int(density*5); i++ {
			if rand.Float64() < dropChance && len(xCoords) > 0 {
				x := xCoords[rand.Intn(len(xCoords))]
				length := int(10 + rand.Float64()*dropLength)

				initialPosition := float64(0)
				if reversed {
					initialPosition = float64(maxY)
				} else {
					initialPosition = float64(minY)
				}

				p.drops = append(p.drops, matrixDrop{
					x:        x,
					length:   length,
					speed:    20 + rand.Float64()*30,
					position: initialPosition,
					bright:   0.8 + rand.Float64()*0.2,
				})
			}
		}
	}
}

func (p *MatrixPattern) applyColor(pixelIdx int, brightness float64, x, y int16) {
	maskColor := p.GetColorMask().GetColorAt(Point{x, y})
	(*p.pixelMap.pixels)[pixelIdx].color = Color{
		R: colorPigment(float64(maskColor.R) * brightness),
		G: colorPigment(float64(maskColor.G) * brightness),
		B: colorPigment(float64(maskColor.B) * brightness),
		W: 0,
	}
}

func (p *MatrixPattern) drawDrops(pixelLookup map[int16]map[int16]int, reversed bool) {
	for _, drop := range p.drops {
		// skip drops with invalid X coordinates
		if _, exists := pixelLookup[drop.x]; !exists {
			continue
		}

		headY := int16(drop.position)

		// tail position depends on direction
		var tailY int16
		if reversed {
			tailY = headY + int16(drop.length)
		} else {
			tailY = headY - int16(drop.length)
		}

		// determine y range to iterate through
		var startY, endY int16
		if reversed {
			startY, endY = headY, tailY
		} else {
			startY, endY = tailY, headY
		}

		p.drawDropSegments(pixelLookup, drop, headY, startY, endY, reversed)
	}
}

func (p *MatrixPattern) drawDropSegments(pixelLookup map[int16]map[int16]int, drop matrixDrop,
	headY, startY, endY int16, reversed bool) {

	if columnMap, exists := pixelLookup[drop.x]; exists {
		for y := startY; y <= endY; y++ {
			// calculate brightness - head is brightest, tail fades out
			var distFromHead float64
			if reversed {
				distFromHead = float64(y - headY)
			} else {
				distFromHead = float64(headY - y)
			}
			brightness := math.Max(0, 1.0-distFromHead/float64(drop.length)) * drop.bright

			if pixelIdx, hasPixel := columnMap[y]; hasPixel {
				p.applyColor(pixelIdx, brightness, drop.x, y)
			}

			p.drawNearbyPixels(pixelLookup, drop, y, headY, brightness, reversed)
		}
	}
}

func (p *MatrixPattern) drawNearbyPixels(pixelLookup map[int16]map[int16]int, drop matrixDrop,
	y, headY int16, centerBrightness float64, reversed bool) {

	for dx := int16(-2); dx <= 2; dx++ {
		if dx == 0 {
			continue // already handled the exact position
		}

		if nearbyColumn, hasColumn := pixelLookup[drop.x+dx]; hasColumn {
			if pixelIdx, hasPixel := nearbyColumn[y]; hasPixel {
				// fade by horizontal distance
				distFactor := 1.0 - math.Abs(float64(dx))/3.0
				brightness := centerBrightness * distFactor

				// only draw if brightness is significant
				if brightness > 0.1 {
					p.applyColor(pixelIdx, brightness, drop.x+dx, y)
				}
			}
		}
	}
}

func (p *MatrixPattern) addSparkles(density float64) {
	sparkleCount := int(density * 10)
	for i := 0; i < sparkleCount; i++ {
		if len(*p.pixelMap.pixels) > 0 {
			idx := rand.Intn(len(*p.pixelMap.pixels))
			brightness := 0.5 + rand.Float64()*0.5
			pixel := (*p.pixelMap.pixels)[idx]

			p.applyColor(idx, brightness, pixel.x, pixel.y)
		}
	}
}

func (p *MatrixPattern) GetName() string {
	return "matrix"
}

type MatrixUpdateRequest struct {
	Parameters MatrixParameters `json:"parameters"`
}

func (r *MatrixUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (p *MatrixPattern) GetPatternUpdateRequest() PatternUpdateRequest {
	return &MatrixUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *MatrixPattern) TransitionFrom(source Pattern, progress float64) {
	DefaultTransitionFromPattern(p, source, progress, p.pixelMap)
}
