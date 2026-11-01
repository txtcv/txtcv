// Package cv loads a CV file (cv.yaml or cv.json) into its decoded JSON Resume
// shape, validating it against the embedded schema.
package cv

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	"github.com/txtcv/txtcv/schema"
)

// Load reads path, decodes it by extension (YAML is converted to its JSON
// form first, so both formats yield one structure), and validates the result
// against the embedded JSON Resume schema. The decoded value is the plain
// JSON shape: map[string]any, []any, string, float64, bool, nil.
func Load(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var decoded []byte
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		if decoded, err = yaml.YAMLToJSON(content); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
	case ".json":
		decoded = content
	default:
		return nil, fmt.Errorf("%s: unsupported CV file format: want .yaml or .json", path)
	}

	var cv map[string]any
	if err := json.Unmarshal(decoded, &cv); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if cv == nil {
		return nil, fmt.Errorf("%s is empty: want a JSON Resume object", path)
	}
	if err := schema.Valid(cv); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON Resume: %w", path, err)
	}
	return cv, nil
}
