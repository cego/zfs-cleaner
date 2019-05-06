package conf

import (
	"bufio"
	"strings"
	"testing"
)

func TestPlanLineError(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name: "testplan",
		conf: c,
	}

	cases := []string{"# unterminated plan", "path", "path /with space", "plan {\n", "#}"}
	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))

		ret := p.planLine(s)

		if ret != nil {
			t.Fatalf("%d planLine() did not return nil on syntax error", i)
		}

		if s.err == nil {
			t.Fatalf("%d planLine() failed to detect syntax error", i)
		}
	}
}

func TestKeep(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name: "testplan",
		conf: c,
	}

	cases := []string{"keep 1d for 30d", "keep 1s for 1h", "keep    30s for 30m", "keep 1h for 30d"}
	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		ret := p.keep(s)

		if s.err != nil {
			t.Fatalf("%d keep() returned unexpected error: %s", i, s.err.Error())
		}

		if ret == nil {
			t.Fatalf("%d keep() did not return action", i)
		}
	}
}

func TestKeepError(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name: "testplan",
		conf: c,
	}

	cases := []string{"keep -1d for 30d", "keep 1d for -30d", "keep -1d for -30d", "keep # comment", "keep 1d for 1s", "keep }"}
	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		ret := p.keep(s)

		if s.err == nil {
			t.Fatalf("%d keep() did not return error", i)
		}

		if ret != nil {
			t.Fatalf("%d keep() returned an action", i)
		}
	}
}

func TestKeepLatest(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name: "testplan",
		conf: c,
	}

	cases := []string{"keep latest 10", "keep latest 1", "keep latest 10 // comment", "keep latest 100000"}
	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		ret := p.keepLatest(s)

		if s.err != nil {
			t.Fatalf("%d keepLatest() returned unexpected error: %s", i, s.err.Error())
		}

		if ret == nil {
			t.Fatalf("%d keepLatest() did not return action", i)
		}
	}
}

func TestKeepLatestError(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name: "testplan",
		conf: c,
	}

	cases := []string{"keep latest -10", "keep latest", "keep latest {", "keep latest 0x30", "keep latest 0"}
	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		ret := p.keepLatest(s)

		if s.err == nil {
			t.Fatalf("%d keepLatest() did not return error", i)
		}

		if ret != nil {
			t.Fatalf("%d keepLatest() returned an action", i)
		}
	}
}

func TestPath(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name: "testplan",
		conf: c,
	}

	cases := []string{"path /", "path /path/to/something"}
	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		ret := p.path(s)

		if s.err != nil {
			t.Fatalf("%d path() returned unexpected error: %s", i, s.err.Error())
		}

		if ret == nil {
			t.Fatalf("%d path() did not return action", i)
		}
	}
}

func TestPathError(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name: "testplan",
		conf: c,
	}

	cases := []string{"path with spaces/something"}
	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		ret := p.path(s)

		if s.err == nil {
			t.Fatalf("%d path() did not return error", i)
		}

		if ret != nil {
			t.Fatalf("%d path() returned an action", i)
		}
	}
}
