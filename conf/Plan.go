package conf

import (
	"strconv"
	"time"
)

// Plan is a description of how the cleaner should behave for specific paths.
type Plan struct {
	Name    string
	Paths   []string
	Latest  int
	Periods []Period
	conf    *Config
}

const (
	ErrSyntaxError      = Error("syntax error")
	ErrUnterminatedPlan = Error("unterminated plan")
	ErrFrequencyTooBig  = Error("frequency cannot be bigger than age")
	ErrLatest1          = Error("latest must be at least 1")
	ErrNoPaths          = Error("no paths defined")
	ErrNoKeeps          = Error("no keep periods defined")
)

func (p *Plan) planLine(s *state) action {
	if !s.scanLine() {
		return s.error(ErrUnterminatedPlan)
	}

	if len(s.fields) == 4 && s.fields[0] == keepIdentifier && s.fields[2] == keepFor {
		return p.keep
	}

	if len(s.fields) == 3 && s.fields[0] == keepIdentifier && s.fields[1] == keepLatest {
		return p.keepLatest
	}

	if len(s.fields) == 2 && s.fields[0] == pathIdentifier {
		return p.path
	}

	if len(s.fields) == 1 && s.fields[0] == blockEnd {
		return p.end
	}

	return s.unparsableToken()
}

func (p *Plan) keep(s *state) action {
	if len(s.fields) < 4 {
		return s.error(ErrSyntaxError)
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
		return s.error(ErrFrequencyTooBig)
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
		return s.error(ErrSyntaxError)
	}

	var num int64
	num, s.err = strconv.ParseInt(s.fields[2], 10, 64)

	if s.err != nil {
		return nil
	}

	if num < 1 {
		return s.error(ErrLatest1)
	}

	p.Latest = int(num)

	return p.planLine
}

func (p *Plan) path(s *state) action {
	if len(s.fields) != 2 {
		return s.error(ErrSyntaxError)
	}

	p.Paths = append(p.Paths, s.fields[1])

	return p.planLine
}

func (p *Plan) end(s *state) action {
	if len(p.Paths) == 0 {
		return s.error(ErrNoPaths)
	}

	if len(p.Periods) == 0 && p.Latest == 0 {
		return s.error(ErrNoKeeps)
	}

	c := p.conf
	p.conf = nil
	c.Plans = append(c.Plans, *p)

	return c.rootLine
}
