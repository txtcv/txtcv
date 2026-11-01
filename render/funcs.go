package render

import (
	"fmt"
	"strings"
	"time"
)

// formatDate renders a JSON Resume partial date for display, accepting the
// schema's accepted forms (2006, 2006-01, 2006-01-01); anything else passes
// through unchanged rather than erroring mid-render.
func formatDate(v any) string {
	s, ok := v.(string)
	if !ok || s == "" {
		return ""
	}
	layouts := []struct{ parse, display string }{
		{"2006-01-02", "Jan 2, 2006"},
		{"2006-01", "Jan 2006"},
		{"2006", "2006"},
	}
	for _, l := range layouts {
		if t, err := time.Parse(l.parse, s); err == nil {
			return t.Format(l.display)
		}
	}
	return s
}

// dateRange renders "Jun 2019 – Sep 2021", with "Present" standing in for a
// missing end date on ongoing entries.
func dateRange(start, end any) string {
	s, e := formatDate(start), formatDate(end)
	switch {
	case s == "" && e == "":
		return ""
	case e == "":
		return s + " – Present"
	case s == "":
		return e
	default:
		return s + " – " + e
	}
}

// joinKeywords flattens a keyword or roles list for inline display.
func joinKeywords(v any) string {
	items, ok := v.([]any)
	if !ok {
		return fmt.Sprint(v)
	}
	strs := make([]string, 0, len(items))
	for _, item := range items {
		if s := fmt.Sprint(item); s != "" {
			strs = append(strs, s)
		}
	}
	return strings.Join(strs, ", ")
}
