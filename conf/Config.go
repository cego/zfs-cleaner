package conf

import (
	"bufio"
	"io"
)

// Config is the top-level configuration for zfs-cleaner.
type Config struct {
	Plans []Plan
}

// Read will read a configuration from r.
func (c *Config) Read(r io.Reader) error {
	s := &state{}
	s.scanner = bufio.NewScanner(r)

	for a := c.rootLine; a != nil; a = a(s) {
	}

	return s.err
}

// rootLine will read a line in the root scope of the configuration file.
func (c *Config) rootLine(s *state) action {
	if !s.scanLine() {
		return nil
	}

	if len(s.fields) == 3 && s.fields[0] == planIdentifier && s.fields[2] == blockStart {
		plan := &Plan{
			Name:   s.fields[1],
			conf:   c,
			Latest: 1,
		}

		return plan.planLine
	}

	return s.unparsableToken()
}
