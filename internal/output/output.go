// Package output provides a shared formatter for --output text|json|yaml.
package output

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Format is the value of the --output flag.
type Format string

const (
	Text Format = "text"
	JSON Format = "json"
	YAML Format = "yaml"
)

// Valid returns true if f is a recognised format.
func (f Format) Valid() bool {
	switch f {
	case Text, JSON, YAML:
		return true
	}
	return false
}

// Print writes v to w in the requested format.
// For Text, v must implement fmt.Stringer; if it doesn't, JSON is used.
// For JSON/YAML, v is marshalled directly.
func Print(w io.Writer, format Format, v any) error {
	switch format {
	case JSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case YAML:
		return yaml.NewEncoder(w).Encode(v)
	default:
		if s, ok := v.(fmt.Stringer); ok {
			_, err := fmt.Fprintln(w, s.String())
			return err
		}
		// fall back to JSON for structs that have no String()
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
}
