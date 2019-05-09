package zfs

import (
	"testing"
	"time"
)

var (
	s0 = Snapshot{"s0", time.Unix(0, 0), false}
	s1 = Snapshot{"s1", time.Unix(1491918988, 0), false}
	s2 = Snapshot{"s2", time.Unix(1491918990, 0), false}
	s3 = Snapshot{"s3", time.Unix(1491919188, 0), false}
)

func TestNewSnapshotFromLine(t *testing.T) {
	cases := []struct {
		input    string
		expected *Snapshot
		err      error
	}{
		{"", nil, ErrMalformedLine},
		{"a", nil, ErrMalformedLine},
		{"s1 1491918988", &s1, nil},
		{"s0 0", &s0, nil},
		{"s1 -1", nil, ErrMalformedLine},
		{"non integer", nil, ErrMalformedLine},
		{"too many fields", nil, ErrMalformedLine},
	}

	for i, c := range cases {
		s, err := NewSnapshotFromLine(c.input)
		if err != nil && c.err == nil {
			t.Fatalf("%d NewSnapshotFromLine() returned unexpected error: %s", i, err.Error())
		}

		if c.expected == nil {
			continue
		}

		if s.Name != c.expected.Name {
			t.Fatalf("%d Got wrong name from '%s', expected '%s', got '%s'", i, c.input, c.expected.Name, s.Name)
		}

		if s.Creation != c.expected.Creation {
			t.Fatalf("%d Got wrong creation from '%s', expected '%s', got '%s'", i, c.input, c.expected.Creation, s.Creation)
		}
	}
}

func TestSnapshotName(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"path@s1", "s1"},
		{"s1", "s1"},
		{"@s1", "s1"},
		{"@", ""},
	}

	for i, c := range cases {
		s := &Snapshot{
			Name: c.input,
		}
		result := s.SnapshotName()
		if result != c.expected {
			t.Fatalf("%d Got wrong name from '%s', expected '%s', got '%s'", i, c.input, c.expected, result)
		}
	}
}
