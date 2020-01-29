package conf

import (
	"bufio"
	"fmt"
	"strings"
)

type (
	// action is a function that will parse the next tokens from the
	// configuration.
	action func(*state) action

	// state will keep state of the configuration file reader.
	state struct {
		scanner *bufio.Scanner
		fields  []string
		err     error
	}
)

// trim will process a line removing comments and space.
func trim(s string) string {
	i := strings.IndexRune(s, commentBashStyle)
	if i >= 0 {
		s = s[0:i]
	}

	i = strings.Index(s, commentCStyle)
	if i >= 0 {
		s = s[0:i]
	}

	s = strings.TrimSpace(s)

	return s
}

// scanLine reads a line from the configuration. If there's no more lines,
// false will be returned.
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

func (s *state) error(err error) action {
	s.err = err

	return nil
}

func (s *state) unparsableToken() action {
	return s.error(fmt.Errorf("unparseable tokens: %v", s.fields))
}
