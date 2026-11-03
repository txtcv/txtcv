package render

import "fmt"

// Page describes the paper geometry used by the browser's print CSS.
// The zero value is A4.
type Page struct {
	// Paper is the paper name: "a4" or "letter". "" means A4.
	Paper string
}

// Default margins, in inches: approximately 14mm top/bottom, 15mm left/right.
const marginCSS = "0.55in 0.59in"

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

// SizeCSS is the CSS @page size keyword for the paper.
// The zero value and any invalid paper fall back to A4.
func (p Page) SizeCSS() string {
	if p.Paper == "letter" {
		return "letter"
	}
	return "a4"
}

// MarginCSS is the CSS shorthand for the default margins.
func (p Page) MarginCSS() string { return marginCSS }
