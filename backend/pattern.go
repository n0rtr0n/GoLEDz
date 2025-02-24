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
	GetLabel() string
	UpdateParameters(AdjustableParameters) error
	GetPatternUpdateRequest() PatternUpdateRequest
	TransitionFrom(source Pattern, progress float64)
	SetColorMask(mask ColorMaskPattern)
	GetColorMask() ColorMaskPattern
}

// BasePattern provides common functionality for all patterns
type BasePattern struct {
	colorMask ColorMaskPattern
	Label     string `json:"label,omitempty"`
}

func (p *BasePattern) SetColorMask(mask ColorMaskPattern) {
	p.colorMask = mask
}

func (p *BasePattern) GetColorMask() ColorMaskPattern {
	return p.colorMask
}

func (p *BasePattern) GetLabel() string {
	return p.Label
}
