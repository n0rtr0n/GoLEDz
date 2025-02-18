package main

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
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

	// Always update current pattern
	if m.currentPattern != nil {
		m.currentPattern.Update()
	}

	// Check if it's time for a new pattern (after interval X)
	if !m.inTransition && time.Since(m.lastPatternSwitch).Seconds() >= m.Parameters.SwitchInterval.Value {
		fmt.Printf("Starting new pattern transition. Current: %v\n", m.currentPattern.GetName())
		m.switchToRandomPattern()
		if m.targetPattern != nil {
			fmt.Printf("Selected target pattern: %v\n", m.targetPattern.GetName())
			m.inTransition = true
			// Send the pattern change request through the controller's channel
			select {
			case m.controller.patternChange <- m.targetPattern:
				fmt.Printf("Started transition to new pattern\n")
			default:
				fmt.Printf("Pattern change channel blocked, skipping transition\n")
				m.inTransition = false
			}
		}
	}
}

func (m *RandomMode) switchToRandomPattern() {
	var availablePatterns []Pattern
	for name, pattern := range m.patterns {
		// Don't include current pattern in available patterns
		if name != m.GetName() && name != "lightsOff" && pattern != m.currentPattern {
			availablePatterns = append(availablePatterns, pattern)
		}
	}

	if len(availablePatterns) > 0 {
		m.targetPattern = availablePatterns[rand.Intn(len(availablePatterns))]
		fmt.Printf("Randomly selected pattern: %v\n", m.targetPattern.GetName())
		m.randomizeParameters()
	}
}

func (m *RandomMode) randomizeParameters() {
	if m.targetPattern == nil {
		return
	}

	// Get the pattern's parameters directly from the pattern interface
	params := reflect.ValueOf(m.targetPattern).Elem().FieldByName("Parameters")
	if !params.IsValid() {
		log.Printf("Pattern %s has no Parameters field", m.targetPattern.GetName())
		return
	}

	// Iterate through all fields in the parameters struct
	for i := 0; i < params.NumField(); i++ {
		field := params.Field(i)
		if !field.CanInterface() {
			continue
		}

		// If the field implements Parameter interface, randomize it
		if param, ok := field.Addr().Interface().(Parameter); ok {
			param.Randomize()
		}
	}
}

// Update TransitionComplete to handle the pattern state updates
func (m *RandomMode) TransitionComplete() {
	fmt.Printf("Transition complete. Moving from %v to %v\n",
		m.currentPattern.GetName(), m.targetPattern.GetName())
	m.currentPattern = m.targetPattern
	m.targetPattern = nil
	m.inTransition = false
	m.lastPatternSwitch = time.Now()
}
