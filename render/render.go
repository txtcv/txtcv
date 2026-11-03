// Package render produces the CV's HTML for browser preview and printing.
package render

import (
	"bytes"
	_ "embed"
	"html/template"
	"strings"
)

//go:embed templates/cv.html
var cvTemplate string

// Params carries everything a render needs.
type Params struct {
	// CV is the decoded JSON Resume document: the plain JSON shape
	// (map[string]any, []any, string, float64, bool, nil).
	CV map[string]any
	// Theme is passed through to the template's data-theme attribute.
	Theme string
	// Page describes the paper the output is designed for. The zero value
	// is A4.
	Page Page
}

// The template is parsed once at package init; a broken template cannot be
// recovered from at runtime, so parse errors become an init panic.
var root = template.Must(
	template.New("cv").Funcs(templateFuncs()).Parse(cvTemplate),
)

// HTML renders the CV for browser preview and printing.
func HTML(p Params) ([]byte, error) {
	var buf bytes.Buffer
	if err := root.Execute(&buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// templateFuncs exposes display helpers to the CV template. All values arrive
// as `any` because the CV is a decoded JSON document.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"date":      formatDate,
		"dateRange": dateRange,
		"join":      joinKeywords,
		"lower":     strings.ToLower,
	}
}
