package zfs

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"time"
)

type (
	// SnapshotList represents a list of snapshots. These will always be sorted
	// by creation time.
	SnapshotList []*Snapshot
)

// NewSnapshotListFromOutput will create a new SnapshotList from the output of
// "zfs list -t snapshot -o name,creation -s creation -H -p".
func NewSnapshotListFromOutput(output []byte, name string) (SnapshotList, error) {
	list := SnapshotList{}
	lastCreation := time.Time{}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		s, err := NewSnapshotFromLine(scanner.Text())
		if err != nil {
			return nil, err
		}

		if !strings.HasPrefix(s.Name, name+"@") {
			continue
		}

		if lastCreation.Sub(s.Creation) > time.Second {
			return nil, fmt.Errorf("output does not appear sorted. %d < %d", s.Creation.Unix(), lastCreation.Unix())
		}

		lastCreation = s.Creation

		list = append(list, s)
	}

	return list, nil
}

// Next will retrieve a pointer to the next Snapshot in l where the snapshot
// creation time is newer than or equal to from.
func (l SnapshotList) Next(from time.Time) *Snapshot {
	for _, snapshot := range l {
		if snapshot.Creation.Sub(from) >= 0 {
			return snapshot
		}
	}

	return nil
}

// Oldest will return a pointer to the oldest snapshot in l.
func (l SnapshotList) Oldest() *Snapshot {
	if len(l) < 1 {
		return nil
	}

	return l[0]
}

// Latest will return a pointer to the latest snapshot in l.
func (l SnapshotList) Latest() *Snapshot {
	if len(l) < 1 {
		return nil
	}

	return l[len(l)-1]
}

// String imeplements Stringer.
func (l SnapshotList) String() string {
	out := "[ "

	for _, s := range l {
		out += fmt.Sprintf("%s ", s.String())
	}

	return out + "]"
}

// KeepLatest will keep the num latest snapshots.
func (l SnapshotList) KeepLatest(num int) {
	start := len(l) - num

	if start < 0 {
		start = 0
	}

	for i := start; i < len(l); i++ {
		l[i].Keep = true
	}
}

// KeepNamed will keep all snapshots named in names.
func (l SnapshotList) KeepNamed(names []string) {
	// Start by indexing names for lookups.
	index := make(map[string]bool)
	for _, name := range names {
		index[name] = true
	}

	for _, snapshot := range l {
		if index[snapshot.SnapshotName()] {
			snapshot.Keep = true
		}
	}
}

// KeepOldest keeps the num oldest snapshots.
func (l SnapshotList) KeepOldest(num int) {
	for i := 0; i < num && i < len(l); i++ {
		l[i].Keep = true
	}
}

// Sieve will mark snapshots to keep according to start time and frequency.
func (l SnapshotList) Sieve(start time.Time, frequency time.Duration) {
	// The ZFS resolution on creation time is one second. If we get a frequency
	// below one second, we have to keep everything after start.
	if frequency < time.Second {
		for _, s := range l {
			if s.Creation.Sub(start) >= 0 {
				s.Keep = true
			}
		}

		return
	}

	for s := l.Next(start); s != nil; s = l.Next(s.Creation.Add(frequency)) {
		s.Keep = true
	}
}

// ResetSieve will mark all snapshots for deletion.
func (l SnapshotList) ResetSieve() {
	for _, s := range l {
		s.Keep = false
	}
}
