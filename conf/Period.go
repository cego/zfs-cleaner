package conf

import (
	"time"
)

// Period is a period for keeping snapshots.
type Period struct {
	Frequency time.Duration
	Age       time.Duration
}
