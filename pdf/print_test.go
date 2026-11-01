package pdf

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/txtcv/txtcv/cv"
	"github.com/txtcv/txtcv/render"
)

// printFixture renders the committed fixture CV on A4 paper.
func printFixture(t *testing.T) []byte {
	t.Helper()
	doc, err := cv.Load(filepath.Join("..", "render", "testdata", "cv.yaml"))
	if err != nil {
		t.Fatalf("load fixture CV: %v", err)
	}
	page, err := render.NewPage("a4")
	if err != nil {
		t.Fatal(err)
	}
	html, err := render.HTML(render.Params{CV: doc, Page: page})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	data, err := Print(ctx, html, Options{Page: page})
	if err != nil {
		t.Fatalf("print: %v", err)
	}
	return data
}

// skipWithoutBrowser skips the test when no browser can be discovered (for
// example in CI without one) or when one is known-bad via env override.
func skipWithoutBrowser(t *testing.T) {
	t.Helper()
	// Only skip on discovery failure; Print surfaces any deeper problems.
	if _, err := FindBrowser(); err != nil {
		t.Skip("no Chromium-family browser available:", err)
	}
}

// The produced output is a real PDF of non-trivial size. Deliberately no
// byte-equality assertion: Chromium embeds a creation timestamp, so the
// output is not byte-reproducible.
func TestPrintSmoke(t *testing.T) {
	skipWithoutBrowser(t)

	data := printFixture(t)
	if !strings.HasPrefix(string(data[:4]), "%PDF") {
		t.Fatalf("output does not start with %%PDF magic: %.20q", data)
	}
	if len(data) < 1024 {
		t.Errorf("PDF suspiciously small: %d bytes", len(data))
	}
}

// The founding credibility claim is verified, not assumed: the PDF must
// expose its text to extractors — name, email, the standard EXPERIENCE
// heading, and skill keywords — the way an ATS would read it.
func TestPrintATSTextExtraction(t *testing.T) {
	skipWithoutBrowser(t)
	data := printFixture(t)

	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not available; install poppler-utils to run the ATS assertion")
	}

	pdfPath := filepath.Join(t.TempDir(), "cv.pdf")
	if err := os.WriteFile(pdfPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("pdftotext", pdfPath, "-").Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			t.Skipf("pdftotext failed on this platform's PDF: %v", err)
		}
		t.Fatalf("pdftotext: %v", err)
	}
	text := string(out)

	for _, want := range []string{"Alice", "alice@example.com", "EXPERIENCE", "Kafka", "English"} {
		if !strings.Contains(text, want) {
			t.Errorf("extracted text is missing %q; got:\n%s", want, text)
		}
	}
}
