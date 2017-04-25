package conf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type (
	Config struct {
		Plans []Plan
	}

	Plan struct {
		Name    string
		Paths   []string
		Latest  int
		Periods []Period
		conf    *Config
	}

	Period struct {
		Frequency time.Duration
		Age       time.Duration
	}

	action func(*state) action

	state struct {
		scanner *bufio.Scanner
		fields  []string
		err     error
	}
)

const (
	planIdentifier = "plan"
	keepIdentifier = "keep"
	pathIdentifier = "path"
)

func trim(s string) string {
	i := strings.IndexRune(s, '#')
	if i >= 0 {
		s = s[0:i]
	}

	i = strings.Index(s, "//")
	if i >= 0 {
		s = s[0:i]
	}

	s = strings.TrimSpace(s)

	return s
}

// Read will read a configuration from r.
func (c *Config) Read(r io.Reader) error {
	s := &state{}
	s.scanner = bufio.NewScanner(r)

	for a := c.rootLine; a != nil; a = a(s) {
	}

	return s.err
}

func (s *state) scanLine() bool {
	s.fields = []string{}

	for s.scanner.Scan() {
		line := s.scanner.Text()
		line = trim(line)
		if line != "" {
			s.fields = strings.Fields(line)
			return true
		}
	}

	return false
}

func (c *Config) rootLine(s *state) action {
	if !s.scanLine() {
		return nil
	}

	if len(s.fields) == 3 && s.fields[0] == planIdentifier && s.fields[2] == "{" {
		plan := &Plan{
			Name:   s.fields[1],
			conf:   c,
			Latest: 1,
		}

		return plan.planLine
	}

	s.err = fmt.Errorf("unparseable tokens: %v", s.fields)
	return nil
}

func (p *Plan) planLine(s *state) action {
	if !s.scanLine() {
		s.err = errors.New("unterminated plan")

		return nil
	}

	if len(s.fields) == 4 && s.fields[0] == keepIdentifier && s.fields[2] == "for" {
		return p.keep
	}

	if len(s.fields) == 3 && s.fields[0] == keepIdentifier && s.fields[1] == "latest" {
		return p.keepLatest
	}

	if len(s.fields) == 2 && s.fields[0] == pathIdentifier {
		return p.path
	}

	if len(s.fields) == 1 && s.fields[0] == "}" {
		return p.end
	}

	s.err = fmt.Errorf("unparseable tokens: %v", s.fields)
	return nil
}

func (p *Plan) keep(s *state) action {
	if len(s.fields) < 4 {
		s.err = errors.New("syntax error")
		return nil
	}

	var frequency time.Duration
	frequency, s.err = parseDuration(s.fields[1])
	if s.err != nil {
		return nil
	}

	var age time.Duration
	age, s.err = parseDuration(s.fields[3])
	if s.err != nil {
		return nil
	}

	if frequency > age {
		s.err = errors.New("frequency cannot be bigger than age")
		return nil
	}

	r := Period{
		Frequency: frequency,
		Age:       age,
	}

	p.Periods = append(p.Periods, r)

	return p.planLine
}

func (p *Plan) keepLatest(s *state) action {
	if len(s.fields) != 3 {
		s.err = errors.New("syntax error")
		return nil
	}

	var num int64
	num, s.err = strconv.ParseInt(s.fields[2], 10, 64)

	if s.err != nil {
		return nil
	}

	if num < 1 {
		s.err = errors.New("latest must be at least 1")
		return nil
	}

	p.Latest = int(num)

	return p.planLine
}

func (p *Plan) path(s *state) action {
	if len(s.fields) != 2 {
		s.err = errors.New("syntax error")
		return nil
	}

	p.Paths = append(p.Paths, s.fields[1])

	return p.planLine
}

func (p *Plan) end(s *state) action {
	if len(p.Paths) == 0 {
		s.err = errors.New("no paths defined")
		return nil
	}

	if len(p.Periods) == 0 && p.Latest == 0 {
		s.err = errors.New("no keep periods defined")
		return nil
	}

	c := p.conf
	p.conf = nil
	c.Plans = append(c.Plans, *p)

	return c.rootLine
}
