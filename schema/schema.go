package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
)

//go:embed resume-schema.json
var resumeSchemaJSON []byte

// resolved lazily parses and resolves the embedded draft-07 JSON Resume
// schema; resolution is required once before validation.
var resolved = sync.OnceValues(func() (*jsonschema.Resolved, error) {
	var s jsonschema.Schema
	if err := json.Unmarshal(resumeSchemaJSON, &s); err != nil {
		return nil, fmt.Errorf("parse resume schema: %w", err)
	}
	r, err := s.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("resolve resume schema: %w", err)
	}
	return r, nil
})

// Valid reports whether instance (a decoded JSON value, e.g. map[string]any
// for the CV object) conforms to the JSON Resume schema.
func Valid(instance any) error {
	r, err := resolved()
	if err != nil {
		return err
	}
	if err := r.Validate(instance); err != nil {
		return fmt.Errorf("not valid against JSON Resume schema: %w", err)
	}
	return nil
}
