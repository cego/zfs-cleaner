package conf

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
	"time"
)

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
				{
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
				{
					Name:   "buh",
					Paths:  []string{"/buh"},
					Latest: 10,
					Periods: []Period{
						{
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
				{
					Name:   "buh",
					Paths:  []string{"/buh"},
					Latest: 10,
					Periods: []Period{
						{
							Frequency: 24 * time.Hour,
							Age:       30 * 24 * time.Hour,
						},
						{
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
protect horse
protect sheep
}`, "", &Config{
			Plans: []Plan{
				{
					Name:   "buh",
					Paths:  []string{"/buh", "/buh/2"},
					Latest: 1,
					Periods: []Period{
						{
							Frequency: 24 * time.Hour,
							Age:       30 * 24 * time.Hour,
						},
						{
							Frequency: time.Hour,
							Age:       24 * time.Hour,
						},
					},
					Protect: []string{
						"horse",
						"sheep",
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
