package slicer

import (
	"time"
)

// CommonLayouts lists frequently used timestamp formats found in log files.
var CommonLayouts = []string{
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02 15:04:05",
	"02/Jan/2006:15:04:05 -0700", // Apache/Nginx combined log
	"Jan  2 15:04:05",            // syslog
}

// AutoParser returns a TimeParser that tries each layout in CommonLayouts
// against the first len(layout) characters of each line.
func AutoParser() TimeParser {
	return func(line string) (time.Time, bool) {
		for _, layout := range CommonLayouts {
			if len(line) < len(layout) {
				continue
			}
			t, err := time.Parse(layout, line[:len(layout)])
			if err == nil {
				return t, true
			}
		}
		return time.Time{}, false
	}
}

// LayoutParser returns a TimeParser that uses the provided Go time layout
// to parse the beginning of each log line.
func LayoutParser(layout string) TimeParser {
	return func(line string) (time.Time, bool) {
		if len(line) < len(layout) {
			return time.Time{}, false
		}
		t, err := time.Parse(layout, line[:len(layout)])
		if err != nil {
			return time.Time{}, false
		}
		return t, true
	}
}
