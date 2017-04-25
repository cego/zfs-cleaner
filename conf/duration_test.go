package conf

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	cases := []struct {
		in  string
		out time.Duration
		err string
	}{
		{"", 0, "duration string too short"},
		{" ", 0, "duration string too short"},
		{"1", 0, "duration string too short"},
		{"12", 0, "unknown unit"},
		{"12k", 0, "unknown unit"},
		{"12year", 0, "unknown unit"},
		{"year", 0, "unknown unit"},
		{"åås", 0, `strconv.ParseInt: parsing "åå": invalid syntax`},
		{"-12h", 0, "negative duration not allowed"},
		{"1h30m", 0, `strconv.ParseInt: parsing "1h30": invalid syntax`},
		{"1s", time.Second, ""},
		{"2m", 2 * time.Minute, ""},
		{"2d", 48 * time.Hour, ""},
		{"1h", time.Hour, ""},
		{"24h", 24 * time.Hour, ""},
		{"0h", 0, ""},
	}

	for i, c := range cases {
		out, err := parseDuration(c.in)

		if err != nil && err.Error() != c.err {
			t.Fatalf("%d Got unexpected error from '%s': expected '%s', got '%s'", i, c.in, c.err, err.Error())
		}

		if err == nil && c.err != "" {
			t.Fatalf("%d Got no error from '%s': expected '%s'", i, c.in, c.err)
		}

		if out != c.out {
			t.Fatalf("%d Got unexpcted output from '%s': expected '%s', got '%s'", i, c.in, c.out, out)
		}
	}
}
