package commands

import (
	"context"
	"testing"

	"github.com/urfave/cli/v3"
)

// An implicit CLI default would mask the theme stored in the CV.
func TestServeThemeFlag(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"use CV metadata", nil, ""},
		{"override CV metadata", []string{"--theme", "creative"}, "creative"},
		{"override with default", []string{"--theme", "default"}, "default"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := serveCommand()
			cmd.Action = func(_ context.Context, cmd *cli.Command) error {
				if got := cmd.String("theme"); got != tc.want {
					t.Errorf("theme override = %q, want %q", got, tc.want)
				}
				return nil
			}
			if err := cmd.Run(context.Background(), append([]string{"serve"}, tc.args...)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
