package main

// PatternMode represents a way of managing/switching between patterns
type PatternMode interface {
	Start()
	Stop()
	Update()
	GetCurrentPattern() Pattern
	SetController(*PixelController)
	UpdateParameters(parameters AdjustableParameters) error
	GetName() string
	GetPatternUpdateRequest() PatternUpdateRequest
}

// BaseModeParameters contains common parameters for modes
type BaseModeParameters struct {
	TransitionEnabled  bool           `json:"transitionEnabled"`
	TransitionDuration FloatParameter `json:"transitionDuration"`
}
