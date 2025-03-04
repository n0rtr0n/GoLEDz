package main

import (
	"encoding/json"
	"time"
)

// OptionType represents the type of an option
type OptionType string

const (
	OPTION_DURATION OptionType = "duration"
	OPTION_BOOLEAN  OptionType = "boolean"
	OPTION_FLOAT    OptionType = "float"
	// Add more types as needed
)

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

// Options holds all configurable settings
type Options struct {
	options            map[string]Option
	TransitionDuration time.Duration `json:"transitionDuration"`
	TransitionEnabled  bool          `json:"transitionEnabled"`
	ActiveMode         string        `json:"activeMode"`
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
	option, exists := o.options[id]
	if !exists {
		return nil, ErrOptionNotFound
	}
	return option, nil
}

// SetOption updates an option's value
func (o *Options) SetOption(id string, value interface{}) error {
	option, err := o.GetOption(id)
	if err != nil {
		return err
	}

	return option.SetValue(value)
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
