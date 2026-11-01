package render

import "fmt"

// Page is the single source of truth for page geometry: it maps a paper name
// to the CSS `@page` size keyword, the dimensions in inches, and the default
// margins. The template and the PDF printer both read it, so they can never
// disagree. The zero value is A4.
type Page struct {
	// Paper is the paper name: "a4" or "letter". "" means A4.
	Paper string
}

var paperSpecs = map[string]paperSpec{
	"a4":     {name: "a4", width: 8.27, height: 11.69},
	"letter": {name: "letter", width: 8.5, height: 11},
}

// paperSpec carries one paper's size, in inches.
type paperSpec struct {
	name   string
	width  float64
	height float64
}

// Default margins, in inches: 14mm top/bottom, 15mm left/right.
const (
	marginTB     = 0.55
	marginLR     = 0.59
	marginCSS    = "0.55in 0.59in"
	defaultPaper = "a4"
)

// NewPage resolves a paper name ("a4" or "letter") to a Page; an empty name
// means A4.
func NewPage(paper string) (Page, error) {
	switch paper {
	case "", "a4", "letter":
		return Page{Paper: paper}, nil
	default:
		return Page{}, fmt.Errorf("unknown paper %q: want \"a4\" or \"letter\"", paper)
	}
}

// paper resolves the page to its paper spec; the zero value and any invalid
// paper fall back to A4.
func (p Page) paper() paperSpec {
	if s, ok := paperSpecs[p.Paper]; ok {
		return s
	}
	return paperSpecs[defaultPaper]
}

// SizeCSS is the CSS @page size keyword for the paper.
func (p Page) SizeCSS() string { return p.paper().name }

// WidthIn is the paper width in inches.
func (p Page) WidthIn() float64 { return p.paper().width }

// HeightIn is the paper height in inches.
func (p Page) HeightIn() float64 { return p.paper().height }

// Margins returns top, right, bottom, left in inches (CSS order).
func (p Page) Margins() (top, right, bottom, left float64) {
	return marginTB, marginLR, marginTB, marginLR
}

// MarginCSS is the CSS shorthand for the default margins.
func (p Page) MarginCSS() string { return marginCSS }
