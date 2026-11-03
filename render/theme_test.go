package render_test

import (
	"strings"
	"testing"

	"github.com/txtcv/txtcv/render"
)

func TestThemePrecedence(t *testing.T) {
	metadata := func(theme any) map[string]any {
		return map[string]any{"meta": map[string]any{"txtcv": map[string]any{"theme": theme}}}
	}
	for _, tc := range []struct {
		name     string
		doc      map[string]any
		override string
		want     string
	}{
		{"no metadata", nil, "", "default"},
		{"CV theme", metadata("code"), "", "code"},
		{"explicit override", metadata("code"), "creative", "creative"},
		{"explicit default", metadata("code"), "default", "default"},
		{"empty CV theme", metadata(""), "", "default"},
		{"non-string CV theme", metadata(123), "", "default"},
		{"unrelated metadata", map[string]any{"meta": map[string]any{"version": "1"}}, "", "default"},
		{"malformed meta", map[string]any{"meta": "invalid"}, "", "default"},
		{"malformed extension", map[string]any{"meta": map[string]any{"txtcv": "invalid"}}, "", "default"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := render.HTML(render.Params{CV: tc.doc, Theme: tc.override})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), `data-theme="`+tc.want+`"`) {
				t.Errorf("expected theme %q on the rendered document", tc.want)
			}
		})
	}
}
