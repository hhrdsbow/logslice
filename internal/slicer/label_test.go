package slicer

import (
	"testing"
)

func TestNewLabeler_DefaultFormat(t *testing.T) {
	l := NewLabeler("", "")
	if l.format != "[%s] %s" {
		t.Fatalf("expected default format, got %q", l.format)
	}
}

func TestLabeler_AddRule_InvalidRegex(t *testing.T) {
	l := NewLabeler("", "")
	err := l.AddRule("[", "bad")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestLabeler_Apply_FirstRuleWins(t *testing.T) {
	l := NewLabeler("", "")
	_ = l.AddRule(`ERROR`, "error")
	_ = l.AddRule(`ERR`, "warn") // would also match but should not fire

	got := l.Apply("ERROR: disk full")
	want := "[error] ERROR: disk full"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLabeler_Apply_NoMatch_NoDefault(t *testing.T) {
	l := NewLabeler("", "")
	_ = l.AddRule(`ERROR`, "error")

	line := "INFO: all good"
	got := l.Apply(line)
	if got != line {
		t.Fatalf("expected line unchanged, got %q", got)
	}
}

func TestLabeler_Apply_NoMatch_WithDefault(t *testing.T) {
	l := NewLabeler("", "info")
	_ = l.AddRule(`ERROR`, "error")

	got := l.Apply("INFO: starting up")
	want := "[info] INFO: starting up"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLabelTransform_Integration(t *testing.T) {
	l := NewLabeler("", "")
	_ = l.AddRule(`WARN`, "warning")

	fn := LabelTransform(l)
	got := fn("WARN: low memory")
	want := "[warning] WARN: low memory"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestStripLabel(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"[error] something bad", "something bad"},
		{"[info] startup", "startup"},
		{"no label here", "no label here"},
		{"[broken", "[broken"},
	}
	for _, c := range cases {
		got := StripLabel(c.input)
		if got != c.want {
			t.Errorf("StripLabel(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
