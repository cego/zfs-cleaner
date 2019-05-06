package conf

import (
	"strconv"
	"time"
)

const (
	ErrDurationTooShort   = Error("duration string too short")
	ErrUnknownUnit        = Error("unknown unit")
	ErrNegativeNotAllowed = Error("negative duration not allowed")
)

func parseDuration(input string) (time.Duration, error) {
	units := map[string]time.Duration{
		"s": time.Second,
		"m": time.Minute,
		"h": time.Hour,
		"d": time.Hour * 24,
		"y": time.Hour * 24 * 365,
	}

	var t time.Duration
	if len(input) < 2 {
		return t, ErrDurationTooShort
	}

	unitIdentifier := input[len(input)-1:]
	unitSize, found := units[unitIdentifier]
	if !found {
		return t, ErrUnknownUnit
	}

	value, err := strconv.ParseInt(input[:len(input)-1], 10, 64)
	if err != nil {
		return t, err
	}

	if value < 0 {
		return t, ErrNegativeNotAllowed
	}

	t = time.Duration(value) * unitSize

	return t, nil
}
