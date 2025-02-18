package main

import (
	"time"
)

// TransitionConfig holds the configuration for pattern transitions
type TransitionConfig struct {
	Duration time.Duration `json:"duration"` // Duration of the transition
	Enabled  bool          `json:"enabled"`  // Whether transitions are enabled
}

// TransitionRequest allows overriding default transition settings in pattern updates
type TransitionRequest struct {
	Transition *TransitionConfig `json:"transition,omitempty"`
}

// TransitionConfigRequest is used for unmarshaling JSON requests
type TransitionConfigRequest struct {
	Duration int  `json:"duration"`
	Enabled  bool `json:"enabled"`
}

// DefaultTransitionFromPattern provides a default implementation for transitioning between patterns
func DefaultTransitionFromPattern(target Pattern, source Pattern, progress float64, pixelMap *PixelMap) {
	// Handle edge cases cleanly
	if progress >= 1.0 {
		// At the end, just run the target pattern directly
		target.Update()
		return
	}

	if progress <= 0.0 {
		// At the start, just run the source pattern directly
		source.Update()
		return
	}

	// Create temporary pixel maps for the transition
	sourcePixels := make([]Pixel, len(*pixelMap.pixels))
	targetPixels := make([]Pixel, len(*pixelMap.pixels))

	// Store original pixel map and pixels
	originalPixels := pixelMap.pixels
	originalPixelsCopy := make([]Pixel, len(*pixelMap.pixels))
	copy(originalPixelsCopy, *pixelMap.pixels)

	// Run source pattern
	pixelMap.pixels = &sourcePixels
	copy(*pixelMap.pixels, originalPixelsCopy)
	source.Update()

	// Run target pattern
	pixelMap.pixels = &targetPixels
	copy(*pixelMap.pixels, originalPixelsCopy)
	target.Update()

	// Restore original pixel map and blend the results
	pixelMap.pixels = originalPixels

	// Linear crossfade between patterns
	for i := range *pixelMap.pixels {
		sourceColor := sourcePixels[i].color
		targetColor := targetPixels[i].color

		(*pixelMap.pixels)[i].color = Color{
			R: colorPigment(float64(sourceColor.R)*(1-progress) + float64(targetColor.R)*progress),
			G: colorPigment(float64(sourceColor.G)*(1-progress) + float64(targetColor.G)*progress),
			B: colorPigment(float64(sourceColor.B)*(1-progress) + float64(targetColor.B)*progress),
		}
	}
}
