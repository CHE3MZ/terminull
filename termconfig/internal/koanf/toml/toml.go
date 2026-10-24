// Package toml adapts github.com/BurntSushi/toml to the koanf.Parser
// interface.
package toml

import (
	burnt "github.com/BurntSushi/toml"
)

// TOML implements a TOML parser.
type TOML struct{}

// Parser returns a TOML parser.
func Parser() *TOML { return &TOML{} }

// Unmarshal parses TOML bytes into a config map.
func (p *TOML) Unmarshal(b []byte) (map[string]any, error) {
	out := map[string]any{}
	if err := burnt.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	normalize(out)
	return out, nil
}

func normalize(v any) any {
	switch value := v.(type) {
	case map[string]any:
		for key, item := range value {
			value[key] = normalize(item)
		}
		return value
	case []map[string]any:
		out := make([]any, len(value))
		for i, item := range value {
			out[i] = normalize(item)
		}
		return out
	case []any:
		for i, item := range value {
			value[i] = normalize(item)
		}
		return value
	default:
		return v
	}
}

// Marshal writes a config map as TOML.
func (p *TOML) Marshal(o map[string]any) ([]byte, error) {
	return burnt.Marshal(o)
}
