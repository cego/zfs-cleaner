package zfs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type (
	// Snapshot represents a sinle ZFS snapshot.
	Snapshot struct {
		Name     string
		Creation time.Time
		Keep     bool
	}
)

var (
	// ErrMalformedLine will be returned if output from zfs is unusable.
	ErrMalformedLine = errors.New("broken line")
)

// NewSnapshotFromLine will try to parse a line from "zfs list" and instantiate
// a new Snapshot.
func NewSnapshotFromLine(line string) (*Snapshot, error) {
	if len(line) < 3 {
		return nil, ErrMalformedLine
	}

	fields := strings.Fields(line)
	if len(fields) != 2 {
		return nil, ErrMalformedLine
	}

	creation, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, err
	}

	if creation < 0 {
		return nil, ErrMalformedLine
	}

	s := Snapshot{
		Name:     fields[0],
		Creation: time.Unix(creation, 0),
	}

	return &s, nil
}

// String implements Stringer.
func (s *Snapshot) String() string {
	return fmt.Sprintf("%s:%d:%v", s.Name, s.Creation.Unix(), s.Keep)
}

// SnapshotName returns the snapshot name part of the full name. This is the
// part after the @.
func (s *Snapshot) SnapshotName() string {
	parts := strings.Split(s.Name, "@")
	if len(parts) != 2 {
		return s.Name
	}

	return parts[1]
}
