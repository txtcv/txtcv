package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"

	"github.com/urfave/cli/v3"

	"github.com/txtcv/txtcv/schema"
)

func seed(format string) ([]byte, error) {
	switch format {
	case "yaml":
		return seedFiles.ReadFile("seed/cv.yaml")
	case "json":
		return seedFiles.ReadFile("seed/cv.json")
	default:
		return nil, fmt.Errorf("unknown seed format %q", format)
	}
}

func initCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "write a sample CV (cv.yaml) with placeholder data to the current directory",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Usage: "file format: yaml or json",
				Value: "yaml",
			},
		},
		Action: initAction,
	}
}

func initAction(ctx context.Context, cmd *cli.Command) error {
	out := cmd.Root().Writer
	if out == nil {
		out = os.Stdout
	}

	format := cmd.String("format")
	switch format {
	case "yaml", "json":
	default:
		return fmt.Errorf("invalid --format %q: want \"yaml\" or \"json\"", format)
	}

	path := "cv." + format
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists; txtcv does not overwrite existing CVs", path)
	}

	content, err := seed(format)
	if err != nil {
		return err
	}
	if err := validateSeed(content, format); err != nil {
		return fmt.Errorf("internal error: bundled %s seed is invalid: %w", format, err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	fmt.Fprintf(out, "Wrote %s: a JSON Resume CV with sample data.\n", path)
	fmt.Fprintf(out, "Replace the placeholder details with your own; whole sections you don't need can be deleted.\n")
	return nil
}

// validateSeed decodes seed content and checks it against the embedded
// JSON Resume schema, so what init writes always validates.
func validateSeed(content []byte, format string) error {
	var instance any
	var err error
	switch format {
	case "json":
		err = json.Unmarshal(content, &instance)
	case "yaml":
		err = yaml.Unmarshal(content, &instance)
	}
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return schema.Valid(instance)
}
