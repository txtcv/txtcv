package render_test

import (
	"fmt"
	"testing"

	"github.com/txtcv/txtcv/render"
)

func TestNewPage(t *testing.T) {
	for _, tc := range []struct {
		paper   string
		wantCSS string
		wantErr bool
	}{
		{"", "a4", false},
		{"a4", "a4", false},
		{"letter", "letter", false},
		{"A4", "", true},
		{"legal", "", true},
	} {
		t.Run(tc.paper, func(t *testing.T) {
			page, err := render.NewPage(tc.paper)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NewPage(%q) should reject the paper", tc.paper)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPage(%q): %v", tc.paper, err)
			}
			if got := page.SizeCSS(); got != tc.wantCSS {
				t.Errorf("SizeCSS() = %q, want %q", got, tc.wantCSS)
			}
		})
	}
}

// The zero value is A4; every accessor agrees with NewPage-resolved A4.
func TestZeroPageIsA4(t *testing.T) {
	a4, _ := render.NewPage("a4")
	var zero render.Page
	if got, want := zero.SizeCSS(), a4.SizeCSS(); got != want {
		t.Errorf("zero Page.SizeCSS() = %q, want %q", got, want)
	}
	if got, want := fmt.Sprint(zero.WidthIn()), fmt.Sprint(a4.WidthIn()); got != want {
		t.Errorf("zero Page.WidthIn() = %s, want %s", got, want)
	}
	if got, want := fmt.Sprint(zero.HeightIn()), fmt.Sprint(a4.HeightIn()); got != want {
		t.Errorf("zero Page.HeightIn() = %s, want %s", got, want)
	}
	if got, want := zero.MarginCSS(), a4.MarginCSS(); got != want {
		t.Errorf("zero Page.MarginCSS() = %q, want %q", got, want)
	}
}

// A4 and Letter, as documented, in inches.
func TestPaperDimensions(t *testing.T) {
	for _, tc := range []struct {
		paper        string
		wantW, wantH float64
	}{
		{"a4", 8.27, 11.69},
		{"letter", 8.5, 11},
	} {
		page, err := render.NewPage(tc.paper)
		if err != nil {
			t.Fatalf("NewPage(%q): %v", tc.paper, err)
		}
		if page.WidthIn() != tc.wantW || page.HeightIn() != tc.wantH {
			t.Errorf("%s = %gx%g in, want %gx%g", tc.paper, page.WidthIn(), page.HeightIn(), tc.wantW, tc.wantH)
		}
	}
}
