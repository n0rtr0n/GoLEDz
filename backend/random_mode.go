package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

type RandomMode struct {
	pixelMap          *PixelMap
	Label             string           `json:"label,omitempty"`
	Parameters        RandomParameters `json:"parameters"`
	patterns          map[string]Pattern
	currentPattern    Pattern
	isActive          bool
	lastPatternSwitch time.Time
	controller        *PixelController
	targetPattern     Pattern
	inTransition      bool
}

type RandomParameters struct {
	BaseModeParameters
	SwitchInterval FloatParameter `json:"switchInterval"`
}

func (m *RandomMode) GetCurrentPattern() Pattern {
	return m.currentPattern
}

func (m *RandomMode) SetController(c *PixelController) {
	m.controller = c
}

func (m *RandomMode) GetName() string {
	return "random"
}

func (m *RandomMode) GetPatternUpdateRequest() PatternUpdateRequest {
	return &RandomUpdateRequest{
		Parameters: m.Parameters,
	}
}

type RandomUpdateRequest struct {
	Parameters RandomParameters `json:"parameters"`
}

func (r *RandomUpdateRequest) GetParameters() AdjustableParameters {
	return r.Parameters
}

func (m *RandomMode) Start() {
	m.isActive = true
	m.lastPatternSwitch = time.Now()
	m.inTransition = false

	// Select and set initial pattern
	m.switchToRandomPattern()
	if m.targetPattern != nil {
		log.Printf("Random mode starting with pattern: %s", m.targetPattern.GetName())
		m.currentPattern = m.targetPattern
		m.currentPattern.Update() // Make sure initial pattern is updated
		m.targetPattern = nil
	}
}

func (m *RandomMode) Stop() {
	m.isActive = false
	if m.currentPattern != nil {
		m.currentPattern = nil
	}
}

func (m *RandomMode) UpdateParameters(parameters AdjustableParameters) error {
	newParams, ok := parameters.(RandomParameters)
	if !ok {
		return fmt.Errorf("could not cast updated parameters for %v mode", m.GetName())
	}
	m.Parameters = newParams
	return nil
}

func (m *RandomMode) Update() {
	if !m.isActive {
		return
	}

	if m.currentPattern != nil {
		m.currentPattern.Update()
	}

	if m.inTransition {
		return
	}

	if time.Since(m.lastPatternSwitch).Seconds() >= m.Parameters.SwitchInterval.Value {
		m.switchToRandomPattern()
		if m.targetPattern != nil {
			go func() {
				m.inTransition = true
				if err := m.controller.SetPattern(m.targetPattern); err != nil {
					m.inTransition = false
					return
				}
				m.currentPattern = m.targetPattern
				m.targetPattern = nil
				m.lastPatternSwitch = time.Now()
			}()
		}
	}
}

func (m *RandomMode) switchToRandomPattern() {
	var availablePatterns []Pattern
	for name, pattern := range m.patterns {
		if name != m.GetName() && name != "lightsOff" {
			availablePatterns = append(availablePatterns, pattern)
		}
	}

	if len(availablePatterns) > 0 {
		m.targetPattern = availablePatterns[rand.Intn(len(availablePatterns))]
		m.randomizeParameters()
	}
}

func (m *RandomMode) randomizeParameters() {
	if m.targetPattern == nil {
		return
	}
	// Copy existing randomization logic from pattern_random.go
}

// Add a method to handle transition completion
func (m *RandomMode) TransitionComplete() {
	m.inTransition = false
}
