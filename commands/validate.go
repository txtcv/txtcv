package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/txtcv/txtcv/cv"
	"github.com/urfave/cli/v3"
)

func validateCommand() *cli.Command {
	return &cli.Command{
		Name:      "validate",
		Usage:     "validate a CV against the bundled JSON Resume schema (defaults to cv.yaml or cv.json)",
		ArgsUsage: "[file]",
		Action:    validateAction,
	}
}

func validateAction(_ context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() > 1 {
		return fmt.Errorf("usage: txtcv validate [file]")
	}
	path := cmd.Args().First()
	if path == "" {
		var err error
		path, err = findCVFile()
		if err != nil {
			return err
		}
	}
	if _, err := cv.Load(path); err != nil {
		return err
	}
	out := cmd.Root().Writer
	if out == nil {
		out = os.Stdout
	}
	_, err := fmt.Fprintf(out, "%s is valid JSON Resume.\n", path)
	return err
}
