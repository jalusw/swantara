package helper

import (
	"fmt"
	"time"
)

func ParseDateStr(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

func ParseDuration(name, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", name, err)
	}
	return duration, nil
}
