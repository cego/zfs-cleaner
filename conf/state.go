package conf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

		previous []*bufio.Scanner
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

			if s.fields[0] == includeIdentifier {
				return s.include()
			}

			return true
		}
	}

	if len(s.previous) > 0 {
		s.scanner = s.previous[0]
		s.previous = s.previous[1:]

		return s.scanLine()
	}

	return false
}

func (s *state) include() bool {
	// This will never err on Unix. Ignore errors.
	paths, _ := filepath.Glob(s.fields[1])

	if len(paths) == 0 {
		s.error(Error(fmt.Sprintf("'%s' matches nothing", s.fields[1])))

		return false
	}

	buffer := bytes.NewBuffer(nil)

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			s.error(err)

			return false
		}

		_, err = io.Copy(buffer, f)
		if err != nil {
			s.error(err)

			return false
		}

		_ = f.Close()

		// Make sure we separate files by newlines in case an
		// included file does not end with newline.
		buffer.WriteString("\n")
	}

	s.previous = append([]*bufio.Scanner{s.scanner}, s.previous...)
	s.scanner = bufio.NewScanner(buffer)

	return s.scanLine()
}

func (s *state) error(err error) action {
	// We only preserve the first error to avoid showing cascading errors.
	if s.err == nil {
		s.err = err
	}

	return nil
}

func (s *state) unparsableToken() action {
	return s.error(fmt.Errorf("unparseable tokens: %v", s.fields))
}
