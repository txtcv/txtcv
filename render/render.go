// Package render produces the CV's HTML for browser preview and printing.
package render

import (
	"bytes"
	"embed"
	"html/template"
	"strings"
)

//go:embed templates/cv.html
var cvTemplate string

//go:embed templates/base.css themes/*.css
var styleFiles embed.FS

// Only trusted, embedded CSS is marked safe for the HTML template. CV content
// still passes through html/template's normal escaping.
var embeddedStyles = func() template.CSS {
	var css strings.Builder
	for _, path := range []string{
		"templates/base.css",
		"themes/default.css",
		"themes/minimal.css",
		"themes/code.css",
		"themes/creative.css",
	} {
		data, err := styleFiles.ReadFile(path)
		if err != nil {
			panic(err)
		}
		css.Write(data)
		css.WriteByte('\n')
	}
	return template.CSS(css.String())
}()

// Params carries everything a render needs.
type Params struct {
	// CV is the decoded JSON Resume document: the plain JSON shape
	// (map[string]any, []any, string, float64, bool, nil).
	CV map[string]any
	// Theme sets the data-theme attribute used by the embedded CSS themes.
	// When empty, meta.txtcv.theme is used, falling back to "default".
	// Unrecognized names use the default styles.
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
	if p.Theme == "" {
		p.Theme = documentTheme(p.CV)
	}
	var buf bytes.Buffer
	if err := root.Execute(&buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// documentTheme reads the vendor extension without requiring optional metadata.
func documentTheme(doc map[string]any) string {
	meta, _ := doc["meta"].(map[string]any)
	txtcv, _ := meta["txtcv"].(map[string]any)
	if theme, ok := txtcv["theme"].(string); ok && theme != "" {
		return theme
	}
	return "default"
}

// templateFuncs exposes display helpers to the CV template. All values arrive
// as `any` because the CV is a decoded JSON document.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"styles":    func() template.CSS { return embeddedStyles },
		"date":      formatDate,
		"dateRange": dateRange,
		"join":      joinKeywords,
		"lower":     strings.ToLower,
	}
}
