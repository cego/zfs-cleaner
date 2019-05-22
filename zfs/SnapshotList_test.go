package zfs

import (
	"fmt"
	"testing"
	"time"
)

var (
	exampleOutput1 = []byte(`playground/fs1@snap1	1491910967
playground/fs1@snap2	1491912383
playground/fs1@snap3	1491913714
playground/fs1@snap4	1491916496
playground/fs1@snap5	1491916503
playground/fs1@snap6	1491916504
playground/fs1@snap7	1491916778
playground/fs1@snap8	1491916780
playground/fs1@snap9	1491916800
playground/fs2@take1	1492990235
playground/fs2@take2	1492990237
playground/fs2@take3	1492990237
`)

	unsortedOutput = []byte(`playground/fs1@snap1	1491910967
playground/fs1@snap2	1491912383
playground/fs1@snap3	1491916503
playground/fs1@snap4	1491913714
playground/fs1@snap5	1491916496
`)

	list = SnapshotList{
		&s1,
		&s2,
		&s3,
	}

	empty = SnapshotList{}
)

func TestNewSnapshotListFromOutput(t *testing.T) {
	s, err := NewSnapshotListFromOutput(exampleOutput1, "playground/fs1")
	if err != nil {
		t.Fatalf("NewSnapshotListFromOutput() errored: %s", err.Error())
	}

	if len(s) != 9 {
		t.Fatalf("NewSnapshotListFromOutput() returned wrong number of snapshots. Got %d, expected %d", len(s), 9)
	}
}

func TestNewSnapshotListFromOutputUnsorted(t *testing.T) {
	_, err := NewSnapshotListFromOutput(unsortedOutput, "playground/fs1")
	if err == nil {
		t.Fatalf("NewSnapshotListFromOutput() did not err on unsorted input")
	}
}

func TestNewSnapshotListFromOutputBroken(t *testing.T) {
	// Test for two many arguments.
	_, err := NewSnapshotListFromOutput([]byte("three argument yay"), "playground/fs1")
	if err == nil {
		t.Fatalf("NewSnapshotListFromOutput() did not err on broken input")
	}

	// Test for a non-integer creation-time.
	_, err = NewSnapshotListFromOutput([]byte("non integer"), "playground/fs1")
	if err == nil {
		t.Fatalf("NewSnapshotListFromOutput() did not err on broken input")
	}
}

func TestSnapshotListNext(t *testing.T) {
	cases := []struct {
		from     int64
		expected *Snapshot
	}{
		{0, &s1},
		{1491918987, &s1},
		{1491918988, &s1},
		{1491919188, &s3},
		{1491919187, &s3},
		{2000000000, nil},
	}

	for _, c := range cases {
		s := list.Next(time.Unix(c.from, 0))
		if s != c.expected {
			t.Fatalf("Next() did not return the expected snapshot")
		}
	}
}

func TestOldest(t *testing.T) {
	oldest := list.Oldest()

	if oldest != &s1 {
		t.Fatalf("Oldest() did not return expected snapshot")
	}
}

func TestOldestEmpty(t *testing.T) {
	oldest := empty.Oldest()

	if oldest != nil {
		t.Fatalf("Oldest() returned a value from an empty list")
	}
}

func TestLatest(t *testing.T) {
	oldest := list.Latest()
	if oldest != &s3 {
		t.Fatalf("Latest() did not return expected snapshot")
	}
}

func TestLatestEmpty(t *testing.T) {
	oldest := empty.Latest()

	if oldest != nil {
		t.Fatalf("Latest() returned a value from an empty list")
	}
}

func TestString(t *testing.T) {
	out := empty.String()

	if out != "[ ]" {
		t.Fatalf("String() returned wrong output for empty list")
	}

	out = list.String()

	if len(out) < 10 {
		t.Fatalf("String() returned wrong output for list")
	}
}

func newSnapshotFromLine(line string) *Snapshot {
	s, err := NewSnapshotFromLine(line)
	if err != nil {
		panic(err.Error())
	}

	return s
}

func testKeep(ii int, t *testing.T, l SnapshotList, expected []bool) {
	for i, e := range expected {
		if e != l[i].Keep {
			fmt.Printf("%s\n", l.String())
			t.Fatalf("%d returned wrong keep for %d, expected %v, got %v", ii, i, e, l[i].Keep)
		}
	}
}

func TestKeepLatest(t *testing.T) {
	cases := []struct {
		num      int
		expected []bool
	}{
		{-1, []bool{false, false, false}},
		{0, []bool{false, false, false}},
		{1, []bool{false, false, true}},
		{2, []bool{false, true, true}},
		{3, []bool{true, true, true}},
		{4, []bool{true, true, true}},
	}

	for ii, c := range cases {
		list.ResetSieve()
		list.KeepLatest(c.num)
		testKeep(ii, t, list, c.expected)
	}
}

func TestKeepOldest(t *testing.T) {
	cases := []struct {
		num      int
		expected []bool
	}{
		{-1, []bool{false, false, false}},
		{0, []bool{false, false, false}},
		{1, []bool{true, false, false}},
		{2, []bool{true, true, false}},
		{3, []bool{true, true, true}},
		{4, []bool{true, true, true}},
	}

	for ii, c := range cases {
		list.ResetSieve()
		list.KeepOldest(c.num)
		testKeep(ii, t, list, c.expected)
	}
}

func TestSieve(t *testing.T) {
	allTrue := []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}
	input := SnapshotList{
		newSnapshotFromLine("name 0"), // 0
		newSnapshotFromLine("name 5"),
		newSnapshotFromLine("name 10"),
		newSnapshotFromLine("name 15"),
		newSnapshotFromLine("name 20"),
		newSnapshotFromLine("name 21"), // 5
		newSnapshotFromLine("name 25"),
		newSnapshotFromLine("name 30"),
		newSnapshotFromLine("name 35"),
		newSnapshotFromLine("name 36"),
		newSnapshotFromLine("name 36"), // 10
		newSnapshotFromLine("name 40"),
		newSnapshotFromLine("name 45"),
		newSnapshotFromLine("name 49"),
		newSnapshotFromLine("name 54"),
		newSnapshotFromLine("name 59"), // 15
		newSnapshotFromLine("name 79"),
	}

	cases := []struct {
		start     time.Time
		frequency time.Duration
		expected  []bool
	}{
		{time.Unix(0, 0), 0, allTrue},
		{time.Unix(0, 0), time.Nanosecond, allTrue},
		{time.Unix(0, 0), time.Second, []bool{
			true, true, true, true, true,
			true, true, true, true, true,
			false, true, true, true, true,
			true, true}},
		{time.Unix(0, 0), 5 * time.Second, []bool{
			true, true, true, true, true,
			false, true, true, true, false,
			false, true, true, false, true,
			true, true}},
		{time.Unix(0, 0), 10 * time.Second, []bool{
			true, false, true, false, true,
			false, false, true, false, false,
			false, true, false, false, true,
			false, true}},
		{time.Unix(0, 0), 100 * time.Second, []bool{
			true, false, false, false, false,
			false, false, false, false, false,
			false, false, false, false, false,
			false, false}},
		{time.Unix(0, 0), 100000000 * time.Second, []bool{
			true, false, false, false, false,
			false, false, false, false, false,
			false, false, false, false, false,
			false, false}},
	}

	for i, c := range cases {
		input.ResetSieve()
		input.Sieve(c.start, c.frequency)

		testKeep(i, t, input, c.expected)
	}
}

func TestKeepNamed(t *testing.T) {
	cases := []struct {
		names    []string
		expected []bool
	}{
		{[]string{"s1"}, []bool{true, false, false}},
		{[]string{"s2"}, []bool{false, true, false}},
		{[]string{"s3"}, []bool{false, false, true}},
		{[]string{"s1", "s3"}, []bool{true, false, true}},
		{[]string{}, []bool{false, false, false}},
		{nil, []bool{false, false, false}},
		{[]string{"nonexisting"}, []bool{false, false, false}},
		{[]string{""}, []bool{false, false, false}},
	}

	for ii, c := range cases {
		list.ResetSieve()
		list.KeepNamed(c.names)
		testKeep(ii, t, list, c.expected)
	}
}
