package slicer

import (
	"testing"
)

func TestColumnExtractor_DefaultWhitespace(t *testing.T) {
	c := NewColumnExtractor([]string{"level", "ts", "msg"})
	cols := c.Extract("INFO 2024-01-01 hello")
	if cols["level"] != "INFO" {
		t.Errorf("expected INFO, got %q", cols["level"])
	}
	if cols["ts"] != "2024-01-01" {
		t.Errorf("expected 2024-01-01, got %q", cols["ts"])
	}
	if cols["msg"] != "hello" {
		t.Errorf("expected hello, got %q", cols["msg"])
	}
}

func TestColumnExtractor_Delimiter(t *testing.T) {
	c := NewColumnExtractor([]string{"a", "b", "c"}, WithDelimiter(","))
	cols := c.Extract("foo,bar,baz")
	if cols["a"] != "foo" || cols["b"] != "bar" || cols["c"] != "baz" {
		t.Errorf("unexpected cols: %v", cols)
	}
}

func TestColumnExtractor_MissingColumns(t *testing.T) {
	c := NewColumnExtractor([]string{"x", "y", "z"})
	cols := c.Extract("only_one")
	if cols["x"] != "only_one" {
		t.Errorf("expected only_one, got %q", cols["x"])
	}
	if cols["y"] != "" || cols["z"] != "" {
		t.Errorf("missing columns should be empty, got %v", cols)
	}
}

func TestColumnExtractor_Regex(t *testing.T) {
	opt, err := WithColumnRegex(`(\w+)\s+(\d+)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := NewColumnExtractor([]string{"name", "count"}, opt)
	cols := c.Extract("errors 42")
	if cols["name"] != "errors" {
		t.Errorf("expected errors, got %q", cols["name"])
	}
	if cols["count"] != "42" {
		t.Errorf("expected 42, got %q", cols["count"])
	}
}

func TestColumnExtractor_InvalidRegex(t *testing.T) {
	_, err := WithColumnRegex(`(`)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestColumnExtractor_RegexNoMatch(t *testing.T) {
	opt, _ := WithColumnRegex(`(\d+)-(\d+)`)
	c := NewColumnExtractor([]string{"a", "b"}, opt)
	cols := c.Extract("no numbers here")
	if cols["a"] != "" || cols["b"] != "" {
		t.Errorf("expected empty cols on no match, got %v", cols)
	}
}

func TestColumnExtractor_Format(t *testing.T) {
	c := NewColumnExtractor([]string{"level", "msg"})
	cols := c.Extract("WARN something went wrong")
	result := c.Format(cols, "[{{level}}] {{msg}}")
	expected := "[WARN] something"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestColumnExtractor_Format_MissingKey(t *testing.T) {
	c := NewColumnExtractor([]string{"level"})
	cols := map[string]string{"level": "ERROR"}
	out := c.Format(cols, "{{level}} {{unknown}}")
	if out != "ERROR {{unknown}}" {
		t.Errorf("unexpected format output: %q", out)
	}
}
