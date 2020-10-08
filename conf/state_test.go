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

func TestEnd(t *testing.T) {
	c := &Config{}
	s := &state{}
	p := &Plan{
		Name:    "testplan",
		Paths:   []string{"/"},
		Periods: []Period{{Frequency: time.Second, Age: time.Hour}},
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
		Periods: []Period{{Frequency: time.Second, Age: time.Hour}},
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

func TestIncludeFile(t *testing.T) {
	s := &state{}

	cases := []string{
		"include /dev/null", // empty file
	}

	for _, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		if s.err != nil {
			t.Errorf("Failed errored on %s", cc)
		}

		s.err = nil
	}
}

func TestIncludeFileError(t *testing.T) {
	s := &state{}

	cases := []string{
		"include /does-not-exists-we-hope", // non existing file
		"include /dev",                     // directory
		"include /etc/shadow",              // insufficient permissions
	}

	for _, cc := range cases {
		s.scanner = bufio.NewScanner(strings.NewReader(cc))
		s.scanLine()

		if s.err == nil {
			t.Errorf("Failed to err on %s", cc)
		}

		s.err = nil
	}
}
