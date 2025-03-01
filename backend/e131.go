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
	// activate each universe separately
	for _, universeNumber := range universeNumbers {
		ch, err := sh.transmitter.Activate(universeNumber)
		if err != nil {
			return fmt.Errorf("failed to activate universe %d: %w", universeNumber, err)
		}

		sh.universes[universeNumber] = ch

		// set destination for each universe
		err = sh.transmitter.SetDestination(universeNumber, controllerAddress, 100)
		if err != nil {
			return fmt.Errorf("failed to set destination for universe %d: %w", universeNumber, err)
		}

		// initialize with zero data, effectively turning off the lights
		zeroData := make([]byte, 512)
		select {
		case ch <- zeroData:
			// successfully sent initialization packet
		default:
			return fmt.Errorf("failed to send initialization packet to universe %d", universeNumber)
		}

		// give the controller time to process each universe
		time.Sleep(25 * time.Millisecond)
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

// verify all universes are active
func (sh *SACNHandler) VerifyUniverses() error {
	// send blank pixels to each universe to verify it's working
	for universeNumber, ch := range sh.universes {
		testData := make([]byte, 512)
		// create a unique pattern for this universe
		for i := range testData {
			testData[i] = byte(universeNumber & 0xFF)
		}

		select {
		case ch <- testData:
			// successfully sent verification data
		default:
			return fmt.Errorf("failed to verify universe %d", universeNumber)
		}

		time.Sleep(25 * time.Millisecond)
	}

	return nil
}

// force synchronization across universes
func (sh *SACNHandler) SyncUniverses() error {
	// some controllers need explicit synchronization
	for universeNumber, ch := range sh.universes {
		// get current data and resend it
		// this forces a refresh of all universes
		currentData := make([]byte, 512)

		select {
		case ch <- currentData:
			// successfully sent sync packet
		default:
			return fmt.Errorf("failed to send sync packet to universe %d", universeNumber)
		}
	}
	return nil
}
