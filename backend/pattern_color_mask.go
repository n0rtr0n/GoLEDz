package main

// defines the interface for patterns that provide color values
type ColorMaskPattern interface {
	Pattern
	// returns the color for a given point in space
	GetColorAt(point Point) Color

	// advances the pattern's internal state (e.g., for animations)
	Update()

	// standard pattern interface methods
	GetName() string
	UpdateParameters(AdjustableParameters) error
	GetPatternUpdateRequest() PatternUpdateRequest
	TransitionFrom(source Pattern, progress float64)
}

// ColorMaskParameters will be embedded in all color mask pattern parameter structs
type ColorMaskParameters struct {
	Speed    FloatParameter   `json:"speed"`
	Reversed BooleanParameter `json:"reversed"`
}
