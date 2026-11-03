package render_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/txtcv/txtcv/cv"
	"github.com/txtcv/txtcv/render"
)

// update, set with `go test ./render -update`, rewrites the golden files.
var update = flag.Bool("update", false, "rewrite golden files")

func loadFixture(t *testing.T) map[string]any {
	t.Helper()
	doc, err := cv.Load(filepath.Join("testdata", "cv.yaml"))
	if err != nil {
		t.Fatalf("load fixture CV: %v", err)
	}
	return doc
}

// Every embedded theme is rendered on both supported paper sizes;
// drift between render and the committed snapshots fails CI.
func TestGolden(t *testing.T) {
	doc := loadFixture(t)

	for _, theme := range []string{"default", "minimal", "code", "creative"} {
		for _, paper := range []string{"a4", "letter"} {
			t.Run(theme+"/"+paper, func(t *testing.T) {
				page, err := render.NewPage(paper)
				if err != nil {
					t.Fatalf("NewPage(%q): %v", paper, err)
				}
				got, err := render.HTML(render.Params{CV: doc, Theme: theme, Page: page})
				if err != nil {
					t.Fatalf("render: %v", err)
				}
				name := paper + ".html"
				if theme != "default" {
					name = theme + "-" + name
				}
				compareGolden(t, filepath.Join("testdata", "golden", name), got)
			})
		}
	}
}

func compareGolden(t *testing.T, path string, got []byte) {
	t.Helper()
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("golden file %s missing: run `go test ./render -update` to create it", path)
		}
		t.Fatalf("read golden: %v", err)
	}
	if !bytes.Equal(bytes.TrimSpace(want), bytes.TrimSpace(got)) {
		t.Errorf("rendered HTML drifted from golden %s; if the change is intended, run `go test ./render -update`", path)
	}
}
