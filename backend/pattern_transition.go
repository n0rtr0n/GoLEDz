package main

import (
	"time"
)

// holds the configuration for pattern transitions
type TransitionConfig struct {
	Duration time.Duration `json:"duration"` // duration of the transition
	Enabled  bool          `json:"enabled"`  // whether transitions are enabled
}

// allows overriding default transition settings in pattern updates
type TransitionRequest struct {
	Transition *TransitionConfig `json:"transition,omitempty"`
}

type TransitionConfigRequest struct {
	Duration int  `json:"duration"`
	Enabled  bool `json:"enabled"`
}

// provides a default implementation for transitioning between patterns
func DefaultTransitionFromPattern(target Pattern, source Pattern, progress float64, pixelMap *PixelMap) {
	// don't update either pattern during parameter transitions
	// only update during pattern changes
	if source.GetName() != target.GetName() {
		target.Update()
	}

	// if we're done transitioning, no need to blend
	if progress >= 1.0 {
		return
	}

	// store current pixels for blending
	currentPixels := make([]Pixel, len(*pixelMap.pixels))
	copy(currentPixels, *pixelMap.pixels)

	// only update source if it's a different pattern
	if source.GetName() != target.GetName() {
		source.Update()
	}
	sourcePixels := make([]Pixel, len(*pixelMap.pixels))
	copy(sourcePixels, *pixelMap.pixels)

	// restore target pattern's pixels
	copy(*pixelMap.pixels, currentPixels)

	// blend between source and target
	for i := range *pixelMap.pixels {
		(*pixelMap.pixels)[i].color = blendColors(sourcePixels[i].color, currentPixels[i].color, progress)
	}
}

func blendColors(c1, c2 Color, progress float64) Color {
	return Color{
		R: colorPigment(float64(c1.R)*(1-progress) + float64(c2.R)*progress),
		G: colorPigment(float64(c1.G)*(1-progress) + float64(c2.G)*progress),
		B: colorPigment(float64(c1.B)*(1-progress) + float64(c2.B)*progress),
	}
}
