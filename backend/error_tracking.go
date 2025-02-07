package main

import (
	"sync"
	"time"
)

// sort of an error rate limiter. helps keeps track of errors we receive
type ErrorTracker struct {
	errors     int
	lastReset  time.Time
	backoff    bool
	mutex      sync.Mutex
	windowSize time.Duration
	maxErrors  int
}

func NewErrorTracker(windowSize time.Duration, maxErrors int) *ErrorTracker {
	return &ErrorTracker{
		lastReset:  time.Now(),
		windowSize: windowSize,
		maxErrors:  maxErrors,
	}
}

func (errorTracker *ErrorTracker) TrackError() bool {
	errorTracker.mutex.Lock()
	defer errorTracker.mutex.Unlock()

	now := time.Now()

	if now.Sub(errorTracker.lastReset) > errorTracker.windowSize {
		errorTracker.errors = 0
		errorTracker.lastReset = now
		errorTracker.backoff = false
	}

	errorTracker.errors++

	if errorTracker.errors >= errorTracker.maxErrors && !errorTracker.backoff {
		errorTracker.backoff = true
		return true
	}

	return errorTracker.backoff
}

func (errorTracker *ErrorTracker) ResetErrors() {
	errorTracker.mutex.Lock()
	defer errorTracker.mutex.Unlock()

	errorTracker.errors = 0
	errorTracker.lastReset = time.Now()
	errorTracker.backoff = false
}

func (errorTracker *ErrorTracker) IsInBackoff() bool {
	errorTracker.mutex.Lock()
	defer errorTracker.mutex.Unlock()
	return errorTracker.backoff
}
