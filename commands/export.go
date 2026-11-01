package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/urfave/cli/v3"

	"github.com/txtcv/txtcv/cv"
	"github.com/txtcv/txtcv/pdf"
	"github.com/txtcv/txtcv/render"
)

func exportCommand() *cli.Command {
	return &cli.Command{
		Name:      "export",
		Usage:     "render the CV to an unwatermarked PDF, locally",
		ArgsUsage: "[file]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Usage: "output format: pdf",
				Value: "pdf",
			},
			&cli.StringFlag{
				Name:  "out",
				Usage: "output path (default: <firstname-lastname>-cv.pdf)",
			},
			&cli.StringFlag{
				Name:  "theme",
				Usage: "theme used to render the CV",
				Value: "default",
			},
			&cli.StringFlag{
				Name:  "paper",
				Usage: "paper size: a4 or letter (default: meta.txtcv.page.size, then a4)",
			},
		},
		Action: exportAction,
	}
}

func exportAction(ctx context.Context, cmd *cli.Command) error {
	out := cmd.Root().Writer
	if out == nil {
		out = os.Stdout
	}

	// HTML/MD/TXT slot in later without flag-surface changes.
	format := cmd.String("format")
	if format != "pdf" {
		return fmt.Errorf("unsupported --format %q: only \"pdf\" is supported today", format)
	}
	if cmd.Args().Len() > 1 {
		return fmt.Errorf("export takes at most one file argument")
	}

	path := cmd.Args().Get(0)
	if path == "" {
		var err error
		if path, err = findCVFile(); err != nil {
			return err
		}
	}
	doc, err := cv.Load(path)
	if err != nil {
		return err
	}

	page, err := resolvePaper(cmd.String("paper"), doc)
	if err != nil {
		return err
	}

	html, err := render.HTML(render.Params{CV: doc, Theme: cmd.String("theme"), Page: page})
	if err != nil {
		return fmt.Errorf("render %s: %w", path, err)
	}

	target := cmd.String("out")
	if target == "" {
		target = outputName(doc)
	}

	data, err := pdf.Print(ctx, html, pdf.Options{Page: page})
	if err != nil {
		return err
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}

	// Silent overwrite: the output is a regenerable artifact.
	fmt.Fprintf(out, "Wrote %s (%s)\n", target, page.SizeCSS())
	return nil
}

// resolvePaper applies the paper resolution order: --paper flag >
// meta.txtcv.page.size > a4. No schema change is needed — meta already
// allows additional properties.
func resolvePaper(paper string, doc map[string]any) (render.Page, error) {
	if paper == "" {
		paper = metaPageSize(doc)
	}
	if paper == "" {
		paper = "a4"
	}
	return render.NewPage(paper)
}

func metaPageSize(doc map[string]any) string {
	meta, ok := doc["meta"].(map[string]any)
	if !ok {
		return ""
	}
	txtcv, ok := meta["txtcv"].(map[string]any)
	if !ok {
		return ""
	}
	page, ok := txtcv["page"].(map[string]any)
	if !ok {
		return ""
	}
	size, _ := page["size"].(string)
	return size
}

// outputName derives the default export path: <name>-cv.pdf slugified from
// basics.name (recruiters and ATS inbox handling prefer that to cv.pdf),
// falling back to cv.pdf when no name.
func outputName(doc map[string]any) string {
	var name string
	if basics, ok := doc["basics"].(map[string]any); ok {
		name, _ = basics["name"].(string)
	}
	if slug := slugify(name); slug != "" {
		return slug + "-cv.pdf"
	}
	return "cv.pdf"
}

// slugify lowercases and joins letter/digit runs with single dashes,
// dropping everything else ("Jane Doe" → "jane-doe").
func slugify(s string) string {
	var b strings.Builder
	lastDash := true // also trims leading separators
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
