package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/txtcv/txtcv/cv"
	"github.com/txtcv/txtcv/pdf"
)

func writeFixtureCV(t *testing.T) {
	t.Helper()
	content, err := seed("yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("cv.yaml", content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func runExport(t *testing.T, args ...string) error {
	t.Helper()
	full := append([]string{"export"}, args...)
	return exportCommand().Run(context.Background(), full)
}

func TestExportRejectsBadFormat(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// The format check comes first, so no CV file is needed here.
	if err := runExport(t, "--format", "docx"); err == nil {
		t.Fatal("--format docx should fail")
	} else if !strings.Contains(err.Error(), "\"pdf\" is supported") {
		t.Errorf("error should name the supported format: %v", err)
	}
	if _, err := os.Stat("cv.docx"); err == nil {
		t.Errorf("no output file should exist after a rejected --format")
	}
}

func TestExportRejectsBadPaper(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFixtureCV(t)

	for _, paper := range []string{"legal", "a3"} {
		if err := runExport(t, "--paper", paper); err == nil {
			t.Fatalf("--paper %s should fail", paper)
		} else if !strings.Contains(err.Error(), `"a4" or "letter"`) {
			t.Errorf("error should name the accepted papers: %v", err)
		}
	}
}

func TestResolvePaper(t *testing.T) {
	fixture, err := seed("yaml")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cv.yaml"), fixture, 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := cv.Load(filepath.Join(dir, "cv.yaml"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Flag wins over meta.
	page, err := resolvePaper("letter", doc)
	if err != nil {
		t.Fatal(err)
	}
	if page.SizeCSS() != "letter" {
		t.Errorf("flag paper ignored: %q", page.SizeCSS())
	}

	// Default is a4 for the stock fixture (meta sets nothing).
	page, err = resolvePaper("", doc)
	if err != nil {
		t.Fatal(err)
	}
	if page.SizeCSS() != "a4" {
		t.Errorf("default paper: %q, want a4", page.SizeCSS())
	}

	// meta.txtcv.page.size wins over the built-in default.
	doc["meta"] = map[string]any{"txtcv": map[string]any{"page": map[string]any{"size": "letter"}}}
	page, err = resolvePaper("", doc)
	if err != nil {
		t.Fatal(err)
	}
	if page.SizeCSS() != "letter" {
		t.Errorf("meta paper ignored: %q", page.SizeCSS())
	}

	// meta of the wrong shape falls back silently.
	doc["meta"] = map[string]any{"txtcv": map[string]any{"page": "not-a-map"}, "odd": true}
	page, err = resolvePaper("", doc)
	if err != nil {
		t.Fatalf("odd meta should not error: %v", err)
	}
	if page.SizeCSS() != "a4" {
		t.Errorf("odd meta fell back to %q, want a4", page.SizeCSS())
	}
}

func TestSlugify(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"Jane Doe", "jane-doe"},
		{"  Jane   van der  Doe  ", "jane-van-der-doe"},
		{"Jane–Doe", "jane-doe"},                     // typographic dash is a separator
		{"Anna-Marie O'Néill", "anna-marie-o-néill"}, // non-ASCII letters stay
		{"李四", "李四"},
		{"", ""},
		{"  --  ", ""},
	} {
		if got := slugify(tc.in); got != tc.want {
			t.Errorf("slugify(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestOutputName(t *testing.T) {
	named := map[string]any{
		"basics": map[string]any{"name": "Jane Doe"},
	}
	if got := outputName(named); got != "jane-doe-cv.pdf" {
		t.Errorf("outputName(named) = %q, want jane-doe-cv.pdf", got)
	}
	for _, tc := range []map[string]any{
		{"basics": map[string]any{}},             // no name
		{"basics": map[string]any{"name": ""}},   // empty name
		{"basics": map[string]any{"name": "!?"}}, // nothing slugifiable
		{},
	} {
		if got := outputName(tc); got != "cv.pdf" {
			t.Errorf("outputName(%v) = %q, want cv.pdf", tc, got)
		}
	}
}

// The full path through export, including the browser, when one is available.
// Skips gracefully wherever a printer cannot launch, e.g. CI without one.
func TestExportEndToEnd(t *testing.T) {
	if _, err := pdf.FindBrowser(); err != nil {
		t.Skip("no Chromium-family browser available:", err)
	}
	dir := t.TempDir()
	t.Chdir(dir)
	writeFixtureCV(t)

	if err := runExport(t, "--out", "mine.pdf"); err != nil {
		t.Fatalf("export --out: %v", err)
	}
	assertPDF(t, "mine.pdf")

	if err := runExport(t); err != nil {
		t.Fatalf("export default name: %v", err)
	}
	assertPDF(t, "alice-cv.pdf") // fixture basics.name is "Alice"
}

func assertPDF(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(data) < 4 || string(data[:4]) != "%PDF" {
		t.Errorf("%s is not a PDF (%d bytes)", path, len(data))
	}
}
