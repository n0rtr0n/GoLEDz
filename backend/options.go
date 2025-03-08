package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// OptionType represents the type of an option
type OptionType string

const (
	OPTION_DURATION         OptionType = "duration"
	OPTION_BOOLEAN          OptionType = "boolean"
	OPTION_FLOAT            OptionType = "float"
	OPTION_COLOR_CORRECTION OptionType = "colorCorrection"
	OPTION_STRUCT           OptionType = "struct"
	// Add more types as needed
)

// TypedValue represents a value with an explicit type
type TypedValue struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// Option represents a configurable setting
type Option interface {
	GetID() string
	GetLabel() string
	GetType() OptionType
	GetValue() interface{}
	SetValue(value interface{}) error
}

// DurationOption represents a time duration setting
type DurationOption struct {
	ID    string        `json:"id"`
	Label string        `json:"label"`
	Value time.Duration `json:"-"` // Hide the actual time.Duration
	Min   int           `json:"min"`
	Max   int           `json:"max"`
}

func (o *DurationOption) GetID() string {
	return o.ID
}

func (o *DurationOption) GetLabel() string {
	return o.Label
}

func (o *DurationOption) GetType() OptionType {
	return OPTION_DURATION
}

func (o *DurationOption) GetValue() interface{} {
	return int(o.Value.Milliseconds())
}

func (o *DurationOption) SetValue(value interface{}) error {
	if val, ok := value.(float64); ok {
		o.Value = time.Duration(val) * time.Millisecond
		return nil
	}
	return ErrInvalidOptionValue
}

// BooleanOption represents a boolean setting
type BooleanOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value bool   `json:"value"`
}

func (o *BooleanOption) GetID() string {
	return o.ID
}

func (o *BooleanOption) GetLabel() string {
	return o.Label
}

func (o *BooleanOption) GetType() OptionType {
	return OPTION_BOOLEAN
}

func (o *BooleanOption) GetValue() interface{} {
	return o.Value
}

func (o *BooleanOption) SetValue(value interface{}) error {
	if val, ok := value.(bool); ok {
		o.Value = val
		return nil
	}
	return ErrInvalidOptionValue
}

// FloatOption represents a floating point setting
type FloatOption struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
}

func (o *FloatOption) GetID() string {
	return o.ID
}

func (o *FloatOption) GetLabel() string {
	return o.Label
}

func (o *FloatOption) GetType() OptionType {
	return OPTION_FLOAT
}

func (o *FloatOption) GetValue() interface{} {
	return o.Value
}

func (o *FloatOption) SetValue(value interface{}) error {
	if val, ok := value.(float64); ok {
		o.Value = val
		return nil
	}
	return ErrInvalidOptionValue
}

// ColorCorrectionSection represents color correction settings for a specific section
type ColorCorrectionSection struct {
	ID    string     `json:"id"`
	Label string     `json:"label"`
	Red   TypedValue `json:"red"`
	Green TypedValue `json:"green"`
	Blue  TypedValue `json:"blue"`
}

// ColorCorrectionOptions represents all color correction settings
type ColorCorrectionOptions struct {
	Enabled  bool                              `json:"enabled"`
	Gamma    TypedValue                        `json:"gamma"`
	Sections map[string]ColorCorrectionSection `json:"sections"`
}

// ColorCorrectionOption represents the top-level color correction option
type ColorCorrectionOption struct {
	ID    string                 `json:"id"`
	Label string                 `json:"label"`
	Value ColorCorrectionOptions `json:"value"`
}

func (o *ColorCorrectionOption) GetID() string {
	return o.ID
}

func (o *ColorCorrectionOption) GetLabel() string {
	return o.Label
}

func (o *ColorCorrectionOption) GetType() OptionType {
	return OPTION_COLOR_CORRECTION
}

func (o *ColorCorrectionOption) GetValue() interface{} {
	return o.Value
}

func (o *ColorCorrectionOption) SetValue(value interface{}) error {
	// For color correction options, we need to handle partial updates
	if valueMap, ok := value.(map[string]interface{}); ok {
		// Update enabled flag if provided
		if enabled, ok := valueMap["enabled"].(bool); ok {
			o.Value.Enabled = enabled
		}

		// Update gamma if provided
		if gammaMap, ok := valueMap["gamma"].(map[string]interface{}); ok {
			if gammaValue, ok := gammaMap["value"].(float64); ok {
				o.Value.Gamma.Value = gammaValue
			}
		}

		// Update sections if provided
		if sectionsMap, ok := valueMap["sections"].(map[string]interface{}); ok {
			for sectionID, sectionData := range sectionsMap {
				if sectionMap, ok := sectionData.(map[string]interface{}); ok {
					section, exists := o.Value.Sections[sectionID]
					if !exists {
						// Skip sections that don't exist in our configuration
						continue
					}

					// Update red if provided
					if redMap, ok := sectionMap["red"].(map[string]interface{}); ok {
						if redValue, ok := redMap["value"].(float64); ok {
							section.Red.Value = redValue
						}
					}

					// Update green if provided
					if greenMap, ok := sectionMap["green"].(map[string]interface{}); ok {
						if greenValue, ok := greenMap["value"].(float64); ok {
							section.Green.Value = greenValue
						}
					}

					// Update blue if provided
					if blueMap, ok := sectionMap["blue"].(map[string]interface{}); ok {
						if blueValue, ok := blueMap["value"].(float64); ok {
							section.Blue.Value = blueValue
						}
					}

					o.Value.Sections[sectionID] = section
				}
			}
		}

		return nil
	}

	return ErrInvalidOptionValue
}

// StructOption represents a structured setting with nested values
type StructOption struct {
	ID    string      `json:"id"`
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

func (o *StructOption) GetID() string {
	return o.ID
}

func (o *StructOption) GetLabel() string {
	return o.Label
}

func (o *StructOption) GetType() OptionType {
	return OPTION_STRUCT
}

func (o *StructOption) GetValue() interface{} {
	return o.Value
}

func (o *StructOption) SetValue(value interface{}) error {
	// For struct options, we need to handle partial updates
	if valueMap, ok := value.(map[string]interface{}); ok {
		// Handle ColorCorrectionOptions specifically
		if colorCorrection, isColorCorrection := o.Value.(ColorCorrectionOptions); isColorCorrection {
			// Update enabled flag if provided
			if enabled, ok := valueMap["enabled"].(bool); ok {
				colorCorrection.Enabled = enabled
			}

			// Update gamma if provided
			if gammaMap, ok := valueMap["gamma"].(map[string]interface{}); ok {
				if gammaValue, ok := gammaMap["value"].(float64); ok {
					colorCorrection.Gamma.Value = gammaValue
				}
			}

			// Update sections if provided
			if sectionsMap, ok := valueMap["sections"].(map[string]interface{}); ok {
				for sectionID, sectionData := range sectionsMap {
					if sectionMap, ok := sectionData.(map[string]interface{}); ok {
						section, exists := colorCorrection.Sections[sectionID]
						if !exists {
							// Skip sections that don't exist in our configuration
							continue
						}

						// Update red if provided
						if redMap, ok := sectionMap["red"].(map[string]interface{}); ok {
							if redValue, ok := redMap["value"].(float64); ok {
								section.Red.Value = redValue
							}
						}

						// Update green if provided
						if greenMap, ok := sectionMap["green"].(map[string]interface{}); ok {
							if greenValue, ok := greenMap["value"].(float64); ok {
								section.Green.Value = greenValue
							}
						}

						// Update blue if provided
						if blueMap, ok := sectionMap["blue"].(map[string]interface{}); ok {
							if blueValue, ok := blueMap["value"].(float64); ok {
								section.Blue.Value = blueValue
							}
						}

						colorCorrection.Sections[sectionID] = section
					}
				}
			}

			o.Value = colorCorrection
			return nil
		}

		// For other struct types, we could add handling here
		return ErrInvalidOptionValue
	}

	return ErrInvalidOptionValue
}

// Add Randomize method to satisfy Parameter interface
func (o *StructOption) Randomize() {
	// Randomizing a struct option doesn't make sense
	// so we'll leave this as a no-op
}

// Options holds all configurable settings
type Options struct {
	options            map[string]Option
	TransitionDuration time.Duration `json:"transitionDuration"`
	TransitionEnabled  bool          `json:"transitionEnabled"`
	ActiveMode         string        `json:"activeMode"`
	mu                 sync.RWMutex
}

// RegisteredOption is used for JSON serialization
type RegisteredOption struct {
	ID    string      `json:"id"`
	Label string      `json:"label"`
	Type  OptionType  `json:"type"`
	Value interface{} `json:"value"`
	Min   *float64    `json:"min,omitempty"`
	Max   *float64    `json:"max,omitempty"`
}

// Errors
var (
	ErrOptionNotFound     = NewError("option not found")
	ErrInvalidOptionValue = NewError("invalid option value")
)

// MarshalJSON customizes JSON serialization
func (o Options) MarshalJSON() ([]byte, error) {
	registeredOptions := make(map[string]RegisteredOption)

	for id, option := range o.options {
		regOption := RegisteredOption{
			ID:    option.GetID(),
			Label: option.GetLabel(),
			Type:  option.GetType(),
			Value: option.GetValue(),
		}

		// Add min/max for duration options
		if durationOpt, ok := option.(*DurationOption); ok {
			min := float64(durationOpt.Min)
			max := float64(durationOpt.Max)
			regOption.Min = &min
			regOption.Max = &max
		}

		// Add min/max for float options
		if floatOpt, ok := option.(*FloatOption); ok {
			min := floatOpt.Min
			max := floatOpt.Max
			regOption.Min = &min
			regOption.Max = &max
		}

		registeredOptions[id] = regOption
	}

	return json.Marshal(registeredOptions)
}

// GetOption returns an option by ID
func (o *Options) GetOption(id string) (Option, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	option, exists := o.options[id]
	if !exists {
		return nil, ErrOptionNotFound
	}
	return option, nil
}

// SetOption updates the value of an option by ID
func (o *Options) SetOption(id string, value interface{}) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	option, exists := o.options[id]
	if !exists {
		return ErrOptionNotFound
	}

	// Capture the current value before any changes
	var currentValueJSON string
	currentValue := option.GetValue()
	currentValueBytes, _ := json.MarshalIndent(currentValue, "", "  ")
	currentValueJSON = string(currentValueBytes)

	// Log the current and incoming values
	incomingValueBytes, _ := json.MarshalIndent(value, "", "  ")
	log.Printf("Option update for: %s\nCURRENT VALUE:\n%s\nNEW VALUE:\n%s",
		id, currentValueJSON, string(incomingValueBytes))

	// Handle special case for TransitionDuration
	if id == "patternTransitionDuration" {
		if durationMs, ok := value.(float64); ok {
			o.TransitionDuration = time.Duration(durationMs) * time.Millisecond
		} else {
			return ErrInvalidOptionValue
		}
	}

	// Handle special case for TransitionEnabled
	if id == "patternTransitionEnabled" {
		if enabled, ok := value.(bool); ok {
			o.TransitionEnabled = enabled
		} else {
			return ErrInvalidOptionValue
		}
	}

	// Update the option value
	if err := option.SetValue(value); err != nil {
		log.Printf("Error updating option %s: %v", id, err)
		return err
	}

	// Log success message
	log.Printf("Option %s updated successfully", id)

	return nil
}

// DefaultOptions returns an Options struct with default values
func DefaultOptions() *Options {
	options := &Options{
		options:            make(map[string]Option),
		TransitionDuration: 1 * time.Second,
		TransitionEnabled:  true,
		ActiveMode:         "", // Empty string means no mode (direct pattern control)
	}

	// Register default options
	options.options["patternTransitionDuration"] = &DurationOption{
		ID:    "patternTransitionDuration",
		Label: "Pattern Transition Duration",
		Value: 2 * time.Second,
		Min:   0,
		Max:   10000,
	}

	options.options["colorMaskTransitionDuration"] = &DurationOption{
		ID:    "colorMaskTransitionDuration",
		Label: "Color Mask Transition Duration",
		Value: 1 * time.Second,
		Min:   0,
		Max:   10000,
	}

	options.options["patternTransitionEnabled"] = &BooleanOption{
		ID:    "patternTransitionEnabled",
		Label: "Pattern Transition Enabled",
		Value: true,
	}

	options.options["colorMaskTransitionEnabled"] = &BooleanOption{
		ID:    "colorMaskTransitionEnabled",
		Label: "Color Mask Transition Enabled",
		Value: true,
	}

	options.options["brightness"] = &FloatOption{
		ID:    "brightness",
		Label: "Brightness",
		Value: 100.0,
		Min:   0.0,
		Max:   100.0,
	}

	options.options["gamma"] = &FloatOption{
		ID:    "gamma",
		Label: "Gamma Correction",
		Value: 1.0, // Default is 1.0 (no correction)
		Min:   0.2, // Lower values make colors more vivid
		Max:   3.0, // Higher values make colors more muted
	}

	// Note: Color correction options will be added later by AddColorCorrectionOptions
	// based on the sections defined in main.go

	return options
}

// Helper function to create a new error
func NewError(message string) error {
	return &customError{message}
}

type customError struct {
	message string
}

func (e *customError) Error() string {
	return e.message
}

// AddColorCorrectionOptions adds hierarchical color correction options
func (options *Options) AddColorCorrectionOptions(sections map[string]Section) {
	// Create the top-level color correction option
	colorCorrectionOpt := &ColorCorrectionOption{
		ID:    "colorCorrection",
		Label: "Color Correction",
		Value: ColorCorrectionOptions{
			Enabled: true,
			Gamma: TypedValue{
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
			Sections: make(map[string]ColorCorrectionSection),
		},
	}

	// Add each section to the color correction options
	for _, section := range sections {
		colorCorrectionOpt.Value.Sections[section.name] = ColorCorrectionSection{
			ID:    section.name,
			Label: section.label,
			Red: TypedValue{
				Value: 100.0,
				Type:  TYPE_FLOAT,
			},
			Green: TypedValue{
				Value: 100.0,
				Type:  TYPE_FLOAT,
			},
			Blue: TypedValue{
				Value: 100.0,
				Type:  TYPE_FLOAT,
			},
		}
	}

	// Register the color correction option
	options.options["colorCorrection"] = colorCorrectionOpt
}

// ResetToDefaults resets all options to their default values
func (o *Options) ResetToDefaults() {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Store current options for logging
	currentOptions, _ := json.Marshal(o.options)

	// Save the color correction option if it exists
	var colorCorrectionOpt Option
	var hasColorCorrection bool
	if opt, exists := o.options["colorCorrection"]; exists {
		colorCorrectionOpt = opt
		hasColorCorrection = true
	}

	// Create new default options
	defaultOpts := DefaultOptions()

	// Replace current options with defaults
	o.options = defaultOpts.options
	o.TransitionDuration = defaultOpts.TransitionDuration
	o.TransitionEnabled = defaultOpts.TransitionEnabled
	o.ActiveMode = defaultOpts.ActiveMode

	// Restore the color correction option if it existed
	if hasColorCorrection {
		o.options["colorCorrection"] = colorCorrectionOpt
	}

	// Log the reset
	log.Printf("All options reset to defaults. Previous options: %s", string(currentOptions))
}

// ResetColorCorrection resets color correction options to their default values
func (o *Options) ResetColorCorrection(sections map[string]Section) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Store current color correction for logging
	var currentColorCorrection string
	if opt, exists := o.options["colorCorrection"]; exists {
		jsonData, _ := json.Marshal(opt)
		currentColorCorrection = string(jsonData)
	}

	// Remove existing color correction option
	delete(o.options, "colorCorrection")

	// Add default color correction options
	o.AddColorCorrectionOptions(sections)

	// Log the reset
	log.Printf("Color correction reset to defaults. Previous settings: %s", currentColorCorrection)
}
