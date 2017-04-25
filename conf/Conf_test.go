package conf

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTrim(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{"", ""},
		{" ", ""},
		{" \t", ""},
		{" \t\r", ""},
		{"\n     \t\r", ""},
		{"# comment", ""},
		{"#comment", ""},
		{"// comment", ""},
		{"//comment", ""},
		{"  # comment", ""},
		{"  #comment", ""},
		{"  // comment", ""},
		{"  //comment", ""},
		{"////// comment", ""},
		{"////// comment # more", ""},
		{"ident // comment", "ident"},
		{"ident / / comment", "ident / / comment"},
		{"ident # comment", "ident"},
	}

	for i, c := range cases {
		out := trim(c.in)
		if out != c.expected {
			t.Fatalf("%d trim() returned unexpected result, got '%s', expected '%s'", i, out, c.expected)
		}
	}
}

func TestRead(t *testing.T) {
	cases := []struct {
		in  string
		err string
		out *Config
	}{
		{"", "", &Config{}}, // Empty config ok.
		{"\n\n# Hello.\n\t// Comment", "", &Config{}},
		{"\nplan buh {", "unterminated plan", &Config{}},
		{"\nplan buh {\nkeep latest 10\n}\n", "no paths defined", &Config{}},
		{`
# A plan for testing
plan buh {
path /buh
keep latest 10
}`, "", &Config{
			Plans: []Plan{
				Plan{
					Name:   "buh",
					Paths:  []string{"/buh"},
					Latest: 10,
				},
			},
		}},

		{`
# A plan for testing
plan buh {
path /buh
keep 1d for 30d
keep latest 10
}`, "", &Config{
			Plans: []Plan{
				Plan{
					Name:   "buh",
					Paths:  []string{"/buh"},
					Latest: 10,
					Periods: []Period{
						Period{
							Frequency: 24 * time.Hour,
							Age:       30 * 24 * time.Hour,
						},
					},
				},
			},
		}},

		{`
plan buh {

path /buh
keep 1d   for 30d
	# Comment.

	keep 1h for 1d // A comment!
keep latest 10
    }`, "", &Config{
			Plans: []Plan{
				Plan{
					Name:   "buh",
					Paths:  []string{"/buh"},
					Latest: 10,
					Periods: []Period{
						Period{
							Frequency: 24 * time.Hour,
							Age:       30 * 24 * time.Hour,
						},
						Period{
							Frequency: time.Hour,
							Age:       24 * time.Hour,
						},
					},
				},
			},
		}},

		{`
plan buh {
path /buh
path /buh/2
keep 1d for 30d
keep 1h for 1d
}`, "", &Config{
			Plans: []Plan{
				Plan{
					Name:   "buh",
					Paths:  []string{"/buh", "/buh/2"},
					Latest: 1,
					Periods: []Period{
						Period{
							Frequency: 24 * time.Hour,
							Age:       30 * 24 * time.Hour,
						},
						Period{
							Frequency: time.Hour,
							Age:       24 * time.Hour,
						},
					},
				},
			},
		}},
	}

	for i, c := range cases {
		r := strings.NewReader(c.in)
		conf := &Config{}
		err := conf.Read(r)

		if err != nil && err.Error() != c.err {
			t.Fatalf("%d Got unexpected error from '%s': expected '%s', got '%s'", i, c.in, c.err, err.Error())
		}

		if err == nil && c.err != "" {
			t.Fatalf("%d Got no error from '%s': expected '%s'", i, c.in, c.err)
		}

		if !reflect.DeepEqual(conf, c.out) {
			t.Fatalf("%d Did not get expected result from '%s': Got %+v, expected %+v", i, c.in, conf, c.out)
		}
	}
}

func TestScanLine(t *testing.T) {
	const input = "\n #comment\nsomething\n  something more  // comment\n"

	s := state{}
	s.scanner = bufio.NewScanner(strings.NewReader(input))

	cases := []struct {
		ret    bool
		fields []string
	}{
		{true, []string{"something"}},
		{true, []string{"something", "more"}},
		{false, []string{}},
	}

	for i, c := range cases {
		ret := s.scanLine()

		if !reflect.DeepEqual(s.fields, c.fields) {
			t.Fatalf("%d scanLine() returned unexpected result, got '%v', expected '%v'", i, s.fields, c.fields)
		}

		if ret != c.ret {
			t.Fatalf("%d scanLine() returned unexpected result, got '%v', expected '%v'", i, ret, c.ret)
		}
	}
}

func TestRootLineEmpty(t *testing.T) {
	c := Config{}
	s := &state{}
	s.scanner = bufio.NewScanner(strings.NewReader("\n\n\n"))

	ret := c.rootLine(s)
	if ret != nil {
		t.Fatalf("rootLine() did not return nil for empty input")
	}
}

func TestRootLineError(t *testing.T) {
	c := Config{}
	s := &state{}
	cases := []string{"syntax error", "plan", "/", "plan buh { hey", "plan buh \n{\n"}

	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))

		ret := c.rootLine(s)

		if ret != nil {
			t.Fatalf("%d rootLine() did not return nil for empty input", i)
		}

		if s.err == nil {
			t.Fatalf("%d rootLine() failed to set error", i)
		}
	}
}

func TestRootLinePlan(t *testing.T) {
	c := Config{}
	s := &state{}
	cases := []string{"plan buh {", "plan     buh   { ", "plan buh \t{ # comment"}

	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))

		ret := c.rootLine(s)

		if ret == nil {
			t.Fatalf("%d rootLine() did not return reader plan line", i)
		}

		if s.err != nil {
			t.Fatalf("%d rootLine() returned unexpected error: %s", i, s.err.Error())
		}
	}
}

func TestPlanLine(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name: "testplan",
		conf: c,
	}

	cases := []string{"path /uhsduh", "keep latest 10", "keep 1m for 1d", "}", "#uehfuehf\n}"}
	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))

		ret := p.planLine(s)

		if ret == nil {
			t.Fatalf("%d planLine() did not return reader plan line", i)
		}

		if s.err != nil {
			t.Fatalf("%d planLine() returned unexpected error: %s", i, s.err.Error())
		}
	}
}

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

func TestEnd(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name:    "testplan",
		Paths:   []string{"/"},
		Periods: []Period{Period{Frequency: time.Second, Age: time.Hour}},
	}

	cases := []string{"}", "} # comment", "  } "}
	for i, cc := range cases {
		p.conf = c
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		ret := p.end(s)

		if s.err != nil {
			t.Fatalf("%d end() returned unexpected error: %s", i, s.err.Error())
		}

		if ret == nil {
			t.Fatalf("%d end() did not return action", i)
		}
	}
}

func TestEndError(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name:    "testplan",
		conf:    c,
		Periods: []Period{Period{Frequency: time.Second, Age: time.Hour}},
	}

	cases := []string{"}", "} # comment", "  } "}
	for i, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		ret := p.end(s)

		if s.err == nil {
			t.Fatalf("%d end() did not return error", i)
		}

		if ret != nil {
			t.Fatalf("%d end() returned an action", i)
		}

		p.Periods = nil
		p.Paths = []string{"/"}
	}
}
