package main

import "fmt"

type SolidColorMask struct {
	BasePattern
	Parameters SolidColorParameters `json:"parameters"`
	Label      string               `json:"label,omitempty"`
}

func (p *SolidColorMask) GetColorAt(point Point) Color {
	return p.Parameters.Color.Value
}

func (p *SolidColorMask) Update() {
	// no animation needed for solid color
}

func (p *SolidColorMask) GetName() string {
	return "solidColorMask"
}

func (p *SolidColorMask) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(SolidColorParameters)
	if !ok {
		return fmt.Errorf("invalid parameters type for SolidColorMask")
	}
	p.Parameters.Color.Update(newParams.Color.Value)
	return nil
}

func (p *SolidColorMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return &SolidColorUpdateRequest{
		Parameters: p.Parameters,
	}
}

func (p *SolidColorMask) TransitionFrom(source Pattern, progress float64) {
	// Use default transition
	DefaultTransitionFromPattern(p, source, progress, nil)
}
