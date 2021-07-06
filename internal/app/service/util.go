package service

import (
	"errors"
	"time"
)

var (
	ErrorTooEarly = errors.New("too early to start service")
	ErrorExpired  = errors.New("service expired")
)

// IsTimeInRange checks if the current time is in between "start" and "end".
// Convert both time to UTC timezone before comparing.
func IsTimeInRange(start time.Time, end time.Time) error {
	now := time.Now().UTC()

	if now.After(end.UTC()) {
		return ErrorExpired
	}

	if now.Before(start.UTC()) {
		return ErrorTooEarly
	}

	return nil
}
