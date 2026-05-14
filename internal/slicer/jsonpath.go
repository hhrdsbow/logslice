package slicer

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONPathExtractor extracts a value from a JSON log line using a dot-separated key path.
type JSONPathExtractor struct {
	path   []string
	fallback string
}

// JSONPathOption configures a JSONPathExtractor.
type JSONPathOption func(*JSONPathExtractor)

// WithJSONFallback sets the string returned when the path is not found.
func WithJSONFallback(s string) JSONPathOption {
	return func(j *JSONPathExtractor) { j.fallback = s }
}

// NewJSONPathExtractor creates an extractor for the given dot-separated path (e.g. "level" or "http.method").
// Returns an error if path is empty.
func NewJSONPathExtractor(dotPath string, opts ...JSONPathOption) (*JSONPathExtractor, error) {
	if strings.TrimSpace(dotPath) == "" {
		return nil, fmt.Errorf("jsonpath: path must not be empty")
	}
	j := &JSONPathExtractor{
		path:     strings.Split(dotPath, "."),
		fallback: "",
	}
	for _, o := range opts {
		o(j)
	}
	return j, nil
}

// Extract parses line as JSON and traverses the configured path.
// Returns the value as a string, or the fallback if the path is absent or line is not valid JSON.
func (j *JSONPathExtractor) Extract(line string) string {
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(line), &root); err != nil {
		return j.fallback
	}
	var cur interface{} = root
	for _, key := range j.path {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return j.fallback
		}
		cur, ok = m[key]
		if !ok {
			return j.fallback
		}
	}
	switch v := cur.(type) {
	case string:
		return v
	case nil:
		return j.fallback
	default:
		return fmt.Sprintf("%v", v)
	}
}

// TransformFunc returns a TransformFn that replaces the line with its extracted value.
// Useful for chaining with NewTransformer.
func (j *JSONPathExtractor) TransformFunc() TransformFn {
	return func(line string) string {
		return j.Extract(line)
	}
}
