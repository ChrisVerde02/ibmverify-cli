package output

import (
	"bytes"
	"strings"
	"testing"
)

type point struct {
	X int    `json:"x" yaml:"x"`
	Y string `json:"y" yaml:"y"`
}

func TestPrint_JSON(t *testing.T) {
	var buf bytes.Buffer
	p := point{X: 1, Y: "hello"}
	if err := Print(&buf, JSON, p); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"x": 1`) {
		t.Errorf("expected JSON with x:1, got: %s", got)
	}
	if !strings.Contains(got, `"y": "hello"`) {
		t.Errorf("expected JSON with y:hello, got: %s", got)
	}
}

func TestPrint_YAML(t *testing.T) {
	var buf bytes.Buffer
	p := point{X: 2, Y: "world"}
	if err := Print(&buf, YAML, p); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "x: 2") {
		t.Errorf("expected YAML with x:2, got: %s", got)
	}
}

func TestPrint_Text_Stringer(t *testing.T) {
	var buf bytes.Buffer
	// exported field so JSON picks it up in the fallback path
	type tok struct{ V string }
	if err := Print(&buf, Text, tok{V: "abc"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "abc") {
		t.Errorf("expected abc in output, got: %s", buf.String())
	}
}

func TestFormat_Valid(t *testing.T) {
	tests := []struct{ f Format; want bool }{
		{Text, true},
		{JSON, true},
		{YAML, true},
		{"csv", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := tt.f.Valid(); got != tt.want {
			t.Errorf("Format(%q).Valid() = %v, want %v", tt.f, got, tt.want)
		}
	}
}
