package main

import (
	"math"
)

// HSVtoRGB converts HSV color values to RGB
// h: 0-360 degrees
// s: 0-1
// v: 0-1
func HSVtoRGB(h, s, v float64) (float64, float64, float64) {
	if s == 0 {
		return v, v, v
	}

	// Convert hue to 0-6 range
	h = h / 60
	i := float64(int(h))
	f := h - i
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	switch int(i) % 6 {
	case 0:
		return v, t, p
	case 1:
		return q, v, p
	case 2:
		return p, v, t
	case 3:
		return p, q, v
	case 4:
		return t, p, v
	default:
		return v, p, q
	}
}

// RGBtoHSV converts RGB color values to HSV
// r, g, b: 0-1
// Returns h: 0-360 degrees, s: 0-1, v: 0-1
func RGBtoHSV(r, g, b float64) (float64, float64, float64) {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	// Calculate value
	v := max

	// Calculate saturation
	var s float64
	if max != 0 {
		s = delta / max
	} else {
		return 0, 0, v // r = g = b = 0, s = 0, h is undefined
	}

	// Calculate hue
	var h float64
	if delta == 0 {
		return 0, s, v // r = g = b, h is undefined
	}

	switch max {
	case r:
		h = (g - b) / delta
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/delta + 2
	case b:
		h = (r-g)/delta + 4
	}
	h *= 60 // Convert to degrees

	return h, s, v
}

// apply gamma correction to a color value
func applyGamma(value float64, gamma float64) float64 {
	return math.Pow(value/255.0, gamma) * 255.0
}
