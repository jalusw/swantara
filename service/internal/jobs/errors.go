package jobs

import "errors"

var (
	ErrJobNotFound = errors.New("job run not found")
	ErrJobRunning  = errors.New("job already running")
)
