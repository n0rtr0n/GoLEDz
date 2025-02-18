package main

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
