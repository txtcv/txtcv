package commands

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/txtcv/txtcv/cv"
	"github.com/txtcv/txtcv/render"

	"github.com/urfave/cli/v3"
)

func serveCommand() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "preview the CV in a local web browser (reads cv.yaml or cv.json from the current directory)",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Usage: "port for the local HTTP server",
				Value: 9090,
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
	doc, err := cv.Load(path)
	if err != nil {
		return err
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
		// The default Page (A4) keeps the browser's own print-to-PDF on the
		// right paper without any serve flag.
		html, err := render.HTML(render.Params{CV: doc, Theme: theme})
		if err != nil {
			http.Error(w, fmt.Sprintf("render %s: %v", path, err), http.StatusInternalServerError)
			return
		}
		w.Write(html)
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
