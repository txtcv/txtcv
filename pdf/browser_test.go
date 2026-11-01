package pdf

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeGetenv clones a map into an injectable env getter.
func fakeGetenv(env map[string]string) func(string) string {
	return func(key string) string { return env[key] }
}

// writeExecutable creates a mode 0o755 file (and its parent dirs).
func writeExecutable(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fake browser"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// writePlain creates a non-executable file (and its parent dirs).
func writePlain(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not a browser"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// browserName is the executable name discovery looks for on the test
// platform.
func browserName() string {
	switch runtime.GOOS {
	case "darwin":
		return "Google Chrome"
	case "windows":
		return "Chrome.exe"
	default:
		return "google-chrome"
	}
}

// install lays down a browser candidate inside dir, using the platform's own
// layout, and returns the full executable path.
func install(t *testing.T, dir, name string, exec bool) string {
	t.Helper()
	var path string
	switch runtime.GOOS {
	case "darwin":
		path = filepath.Join(dir, name+".app", "Contents", "MacOS", name)
	case "windows":
		path = filepath.Join(dir, "Google", "Chrome", "Application", name)
	default:
		path = filepath.Join(dir, name)
	}
	if exec {
		return writeExecutable(t, path)
	}
	return writePlain(t, path)
}

func TestFindBrowserEnvWins(t *testing.T) {
	// The override is the raw executable path; its name is irrelevant.
	exe := writeExecutable(t, filepath.Join(t.TempDir(), "my-browser"))

	got, err := findBrowser(fakeGetenv(map[string]string{browserEnv: exe}), nil)
	if err != nil {
		t.Fatalf("findBrowser: %v", err)
	}
	if got != exe {
		t.Errorf("findBrowser = %q, want %q", got, exe)
	}
}

// A bad override fails loudly instead of silently falling back.
func TestFindBrowserEnvInvalid(t *testing.T) {
	dir := t.TempDir()
	notAFile := filepath.Join(dir, "missing-browser")
	notExecutable := writePlain(t, filepath.Join(dir, "plain"))

	for _, tc := range []struct{ path, reason string }{
		{notAFile, "does not exist"},
		{notExecutable, "is not executable"},
	} {
		_, err := findBrowser(fakeGetenv(map[string]string{browserEnv: tc.path}), nil)
		if err == nil {
			t.Fatalf("%s=%s should be rejected", browserEnv, tc.path)
		}
		for _, want := range []string{browserEnv, tc.path, tc.reason} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q does not mention %q", err, want)
			}
		}
	}
}

// Discovery searches the injected dirs in order, before the platform
// defaults, using the layout of the test platform, and skips candidates that
// are not executable.
func TestFindBrowserCandidates(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows layout is exercised only on code review; see plans/02 §9")
	}

	first := t.TempDir()
	second := t.TempDir()
	firstPath := install(t, first, browserName(), true)
	secondPath := install(t, second, browserName(), true)
	env := fakeGetenv(map[string]string{})

	t.Run("first dir wins", func(t *testing.T) {
		got, err := findBrowser(env, []string{first, second})
		if err != nil {
			t.Fatalf("findBrowser: %v", err)
		}
		if got != firstPath {
			t.Errorf("findBrowser = %q, want %q", got, firstPath)
		}
	})

	t.Run("non-executable is skipped", func(t *testing.T) {
		plain := t.TempDir()
		install(t, plain, browserName(), false)
		got, err := findBrowser(env, []string{plain, second})
		if err != nil {
			t.Fatalf("findBrowser: %v", err)
		}
		if got != secondPath {
			t.Errorf("findBrowser = %q, want %q", got, secondPath)
		}
	})
}

// With no override, no installed candidates, and empty injected dirs, the
// error lists every remedy.
func TestFindBrowserExhausted(t *testing.T) {
	empty := t.TempDir()
	_, err := findBrowser(fakeGetenv(map[string]string{}), []string{empty})
	if err == nil {
		t.Skip("a real browser is installed on this machine; the exhausted case cannot be reached here")
	}
	for _, want := range []string{
		"no Chromium-based browser found",
		"install Google Chrome",
		browserEnv + "=/path/to/browser",
		"txtcv serve",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}
