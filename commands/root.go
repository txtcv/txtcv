// Package commands implements the txtcv CLI subcommands.
package commands

import (
	"github.com/urfave/cli/v3"
)

// Root assembles the txtcv command tree.
func Root() *cli.Command {
	return &cli.Command{
		Name:     "txtcv",
		Usage:    "plaintext CV builder: your CV is a file, your job hunt is a directory",
		Commands: []*cli.Command{initCommand(), serveCommand(), exportCommand()},
	}
}
