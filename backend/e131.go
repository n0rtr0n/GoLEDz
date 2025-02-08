package main

import (
	"fmt"
	"time"
)

type SACNHandler struct {
	transmitter  *Transmitter
	universes    map[uint16]chan<- []byte
	errorTracker *ErrorTracker
}

func NewSACNHandler() (*SACNHandler, error) {
	errorTracker := NewErrorTracker(5*time.Minute, 50)

	config := TransmitterConfig{
		CID:        [16]byte{0x47, 0x6f, 0x4c, 0x45, 0x44, 0x7a}, // "GoLEDz"
		SourceName: "GoLEDz",
		Priority:   100,
	}

	trans, err := NewTransmitter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create transmitter: %w", err)
	}

	return &SACNHandler{
		transmitter:  trans,
		universes:    make(map[uint16]chan<- []byte),
		errorTracker: errorTracker,
	}, nil
}

func (sh *SACNHandler) Setup(universeNumbers []uint16, controllerAddress string) error {
	for _, universeNumber := range universeNumbers {
		ch, err := sh.transmitter.Activate(universeNumber)
		if err != nil {
			return fmt.Errorf("failed to activate universe %d: %w", universeNumber, err)
		}

		sh.universes[universeNumber] = ch

		// Set the destination for this universe
		err = sh.transmitter.SetDestination(universeNumber, controllerAddress, 100)
		if err != nil {
			return fmt.Errorf("failed to set destination for universe %d: %w", universeNumber, err)
		}
	}
	return nil
}

func (sh *SACNHandler) GetUniverses() map[uint16]chan<- []byte {
	return sh.universes
}

func (sh *SACNHandler) GetErrorTracker() *ErrorTracker {
	return sh.errorTracker
}

func (sh *SACNHandler) Close() error {
	return sh.transmitter.Close()
}
