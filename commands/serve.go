package commands

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/txtcv/txtcv/schema"

	"github.com/goccy/go-yaml"
	"github.com/urfave/cli/v3"
)

//go:embed templates/cv.html
var serveHTMLTemplate string

type serveParams struct {
	Path  string
	CV    map[string]any
	Theme string
}

func serveCommand() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "preview the CV in a local web browser (reads cv.yaml or cv.json from the current directory)",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Usage: "port for the local HTTP server",
				Value: 9000,
			},
			&cli.StringFlag{
				Name:  "theme",
				Usage: "theme used to render the CV",
				Value: "default",
			},
		},
		Action: serveAction,
	}
}

func serveAction(ctx context.Context, cmd *cli.Command) error {
	out := cmd.Root().Writer
	if out == nil {
		out = os.Stdout
	}

	port := cmd.Int("port")
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid --port %d: want 1-65535", port)
	}
	theme := cmd.String("theme")

	path, err := findCVFile()
	if err != nil {
		return err
	}
	cv, err := loadCV(path)
	if err != nil {
		return err
	}

	tmpl, err := template.New("cv").Funcs(templateFuncs()).Parse(serveHTMLTemplate)
	if err != nil {
		return fmt.Errorf("internal error: parse CV template: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Rendering happens per request against the CV loaded at startup,
		// so template bugs surface as plain errors instead of crashing serve.
		if err := tmpl.Execute(w, serveParams{Path: path, CV: cv, Theme: theme}); err != nil {
			http.Error(w, fmt.Sprintf("render %s: %v", path, err), http.StatusInternalServerError)
		}
	})

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("listen on port %d: %w", port, err)
	}

	fmt.Fprintf(out, "Serving %s (theme: %s)\n", path, theme)
	fmt.Fprintf(out, "Open http://localhost:%d — press Ctrl+C to stop\n", port)
	return http.Serve(ln, mux)
}

// findCVFile locates the CV in the current directory, preferring cv.yaml.
func findCVFile() (string, error) {
	for _, name := range []string{"cv.yaml", "cv.json"} {
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("no cv.yaml or cv.json found in the current directory; run `txtcv init` to create one")
}

// loadCV reads path, decodes it by extension (YAML is converted to its JSON
// form first, so both formats yield one structure), and validates the result
// against the embedded JSON Resume schema. The decoded value is the plain
// JSON shape: map[string]any, []any, string, float64, bool, nil.
func loadCV(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var decoded []byte
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		if decoded, err = yaml.YAMLToJSON(content); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
	case ".json":
		decoded = content
	default:
		return nil, fmt.Errorf("%s: unsupported CV file format: want .yaml or .json", path)
	}

	var cv map[string]any
	if err := json.Unmarshal(decoded, &cv); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if cv == nil {
		return nil, fmt.Errorf("%s is empty: want a JSON Resume object", path)
	}
	if err := schema.Valid(cv); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON Resume: %w", path, err)
	}
	return cv, nil
}

// templateFuncs exposes display helpers to the CV template. All values arrive
// as `any` because the CV is a decoded JSON document.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"date":      formatDate,
		"dateRange": dateRange,
		"join":      joinKeywords,
		"lower":     strings.ToLower,
	}
}

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
