package pdf

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/txtcv/txtcv/render"
)

// Options carries the settings of a print.
type Options struct {
	// Page is the paper the print target is designed for — the same struct
	// that produced the HTML's @page rule. The zero value is A4.
	Page render.Page
}

// printTimeout bounds one browser launch plus print.
const printTimeout = 60 * time.Second

// Print renders html to PDF bytes using a locally discovered Chromium-family
// browser. The HTML is written to a temporary directory, reached only through
// a file:// URL, so the browser makes no network calls; its own background
// traffic is disabled in the launch flags.
func Print(ctx context.Context, html []byte, opts Options) ([]byte, error) {
	browser, err := FindBrowser()
	if err != nil {
		return nil, err
	}
	pg := opts.Page

	dir, err := os.MkdirTemp("", "txtcv-pdf-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "cv.html")
	if err := os.WriteFile(file, html, 0o600); err != nil {
		return nil, fmt.Errorf("write %s: %w", file, err)
	}
	u := (&url.URL{Scheme: "file", Path: filepath.ToSlash(file)}).String()

	ctx, cancel := context.WithTimeout(ctx, printTimeout)
	defer cancel()

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocatorOptions(browser)...)
	defer allocCancel()
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	if err := chromedp.Do(browserCtx,
		chromedp.Navigate(u),
		chromedp.WaitReady("body"),
	); err != nil {
		return nil, fmt.Errorf("open the CV in %s: %w", filepath.Base(browser), err)
	}
	// document.fonts.ready: instant with system fonts, future-proofs
	// self-hosted-font themes.
	if _, err := chromedp.Run(browserCtx, chromedp.Poll[bool](
		`document.fonts.status === "loaded"`,
		chromedp.WithPollingTimeout(15*time.Second),
	)); err != nil {
		return nil, fmt.Errorf("wait for fonts in %s: %w", filepath.Base(browser), err)
	}

	data, err := chromedp.Run(browserCtx, chromedp.PrintToPDF(
		chromedp.PDFPaper(chromedp.PaperSize{Width: pg.WidthIn(), Height: pg.HeightIn()}),
		chromedp.PDFMargins(pg.Margins()),
		chromedp.PDFPrintBackground(),
		// The CSS @page rule and these numbers come from one Page struct;
		// whichever wins, the result is identical.
		chromedp.PDFPreferCSSPageSize(),
		// Tagged PDF structure tree + document outline from headings: what
		// makes the PDF structure-aware for ATS parsers and screen readers.
		chromedp.PDFOutlineAndTagged(),
	))
	if err != nil {
		return nil, fmt.Errorf("print PDF: %w", err)
	}
	return data, nil
}

// allocatorOptions launches the discovered browser with the flags export
// needs: headless (explicitly the modern one — legacy headless was removed in
// Chrome 109+; --headless=new is accepted since), and flags that keep the
// browser itself quiet. Defaults also give --no-first-run,
// --no-default-browser-check, --disable-background-networking and
// --hide-scrollbars.
func allocatorOptions(browser string) []chromedp.ExecAllocatorOption {
	return append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browser),
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-component-update", true),
		chromedp.DisableGPU,
	)
}
