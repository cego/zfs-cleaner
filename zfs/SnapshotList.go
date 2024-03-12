package zfs

import (
	"bufio"
	"bytes"
	"fmt"
	"time"
)

type (
	// SnapshotList represents a list of snapshots. These will always be sorted
	// by creation time.
	SnapshotList []*Snapshot
)

// NewSnapshotListFromDataset will create a new SnapshotList from the output of the provided ZfsExecutor
func (l SnapshotList) NewSnapshotListFromDataset(zfsExecutor Executor, dataset string) (SnapshotList, error) {
	list := SnapshotList{}
	lastCreation := time.Time{}
	output, err := zfsExecutor.GetSnapshotList(dataset)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		s, err := NewSnapshotFromLine(scanner.Text())
		if err != nil {
			return nil, err
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

// KeepHolds will keep all snapshots with zfs holds on it.
func (l SnapshotList) KeepHolds(zfsExecutor Executor) error {
	for _, snapshot := range l {
		hasHolds, err := zfsExecutor.HasHolds(snapshot.Name)
		if err != nil {
			return err
		}
		if hasHolds {
			snapshot.Keep = true
		}
	}
	return nil
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

	// We move start back to a point in time where it will "snap" to a
	// frequency based on some predefined reference time in the past.
	// We do this to make zfs-cleaner more resilient to the execution start
	// time of zfs-cleaner itself.
	// Imagine we have two ZFS servers synchronizing snapshots. If they both
	// run zfs-cleaner, they could easily diverge on what snapshots to keep,
	// and which to delete based on execution time.
	// This trick should make sure that we're consistent about what to
	// delete and what to destroy since ZFS snapshots maintain the creation-
	// time across pools.
	// Please note that zfs-cleaner run should happen at roughly the same
	// time on all hosts. As long as zfs-cleaners runs more often than the
	// keep frequency, the result should be predictable.
	// If you have a planline like "keep 2h for 24h", you should execute
	// zfs-cleaner every hour to ensure consistency, but it's not important
	// *when* zfs-cleaner executes, as long as it's in a shorter span than
	// the keep frequency.
	// The reference time is the standard UNIX epoch at midnight January
	// 1st 1970, but that's not important, as long it's guaranteed to be
	// in the past.

	// Calculate how much to move the start back in time to snap to a time
	// aligned to the frequency.
	offset := time.Duration(start.UnixNano()) % frequency

	// Do the actual move. If start was already snapped (by chance) when
	// passed, this does nothing since offset will be zero.
	start = start.Add(-offset)

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
