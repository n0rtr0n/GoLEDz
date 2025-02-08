package main

// controller-specific limitation when not running in expanded mode
const MAX_PIXEL_LENGTH = 340
const MAX_HUE_VALUE = 360
const MAX_DEGREES = 360
const MAX_SATURATION = 1.0

// abitrary for now; we'll calculate this later
const MAX_X_POSITION = 600

// TODO: return and handle any errors encountered in updating patterns

type PatternUpdateRequest interface {
	GetParameters() AdjustableParameters
}

// pattern
type Pattern interface {
	Update()
	GetName() string
	UpdateParameters(AdjustableParameters) error
	GetPatternUpdateRequest() PatternUpdateRequest
}
