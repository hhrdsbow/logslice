package slicer

import (
	"regexp"
	"strings"
)

// SeverityLevel represents a log severity tier.
type SeverityLevel int

const (
	SeverityUnknown SeverityLevel = iota
	SeverityDebug
	SeverityInfo
	SeverityWarn
	SeverityError
	SeverityFatal
)

var severityPatterns = []struct {
	level   SeverityLevel
	pattern *regexp.Regexp
}{
	{SeverityFatal, regexp.MustCompile(`(?i)\b(fatal|panic|critical)\b`)},
	{SeverityError, regexp.MustCompile(`(?i)\b(error|err|exception)\b`)},
	{SeverityWarn, regexp.MustCompile(`(?i)\b(warn(?:ing)?)\b`)},
	{SeverityInfo, regexp.MustCompile(`(?i)\b(info|notice)\b`)},
	{SeverityDebug, regexp.MustCompile(`(?i)\b(debug|trace|verbose)\b`)},
}

// String returns the human-readable name of the severity level.
func (s SeverityLevel) String() string {
	switch s {
	case SeverityDebug:
		return "DEBUG"
	case SeverityInfo:
		return "INFO"
	case SeverityWarn:
		return "WARN"
	case SeverityError:
		return "ERROR"
	case SeverityFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// DetectSeverity returns the highest severity level found in line.
func DetectSeverity(line string) SeverityLevel {
	upper := strings.ToUpper(line)
	_ = upper
	for _, sp := range severityPatterns {
		if sp.pattern.MatchString(line) {
			return sp.level
		}
	}
	return SeverityUnknown
}

// SeverityFilter passes only lines whose detected severity meets the minimum.
type SeverityFilter struct {
	min SeverityLevel
}

// NewSeverityFilter creates a filter that accepts lines at or above min.
func NewSeverityFilter(min SeverityLevel) *SeverityFilter {
	if min < SeverityDebug {
		min = SeverityDebug
	}
	return &SeverityFilter{min: min}
}

// Accept returns true when the line's severity is >= the minimum level.
func (f *SeverityFilter) Accept(line string) bool {
	return DetectSeverity(line) >= f.min
}
