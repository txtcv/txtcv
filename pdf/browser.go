// Package pdf prints the CV's rendered HTML to an unwatermarked,
// ATS-friendly PDF, entirely locally: a discovered Chromium-family browser
// loads a file:// URL and answers over the DevTools protocol on 127.0.0.1.
// There is no server and no network call in the privacy sense.
package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// browserEnv overrides browser discovery with the path to a specific
// executable.
const browserEnv = "TXTCV_BROWSER"

// The error copy when no browser is found. Every remedy is listed, ending in
// the zero-dependency manual fallback.
const noBrowserMsg = "no Chromium-based browser found for PDF export.\n" +
	"  1. install Google Chrome / Chromium / Edge / Brave, or\n" +
	"  2. set " + browserEnv + "=/path/to/browser, or\n" +
	"  3. run `txtcv serve` and print to PDF from the browser (⌘P → Save as PDF)."

// FindBrowser returns the path of a Chromium-family browser executable:
// TXTCV_BROWSER when set, otherwise the first installed candidate in the
// platform's usual locations. An error from FindBrowser is actionable and
// safe to show as-is.
func FindBrowser() (string, error) {
	return findBrowser(os.Getenv, nil)
}

// findBrowser drives discovery. getenv and extraDirs are injectable so unit
// tests exercise the resolution order without a real browser; extraDirs are
// searched before the platform defaults, each in the platform's layout.
func findBrowser(getenv func(string) string, extraDirs []string) (string, error) {
	if path := getenv(browserEnv); path != "" {
		if !executable(path) {
			reason := "does not exist"
			if _, err := os.Stat(path); err == nil {
				reason = "is not executable"
			}
			return "", fmt.Errorf("%s=%s %s: want the path of a Chrome-family browser executable", browserEnv, path, reason)
		}
		return path, nil
	}
	for _, path := range candidatePaths(getenv, extraDirs) {
		if executable(path) {
			return path, nil
		}
	}
	return "", fmt.Errorf(noBrowserMsg)
}

// macApps are the app bundles searched under each macOS applications dir;
// the check resolves …app/Contents/MacOS/<name>.
var macApps = []string{"Google Chrome", "Google Chrome Canary", "Chromium", "Microsoft Edge", "Brave Browser"}

// linuxNames are the executables searched on Linux & other Unix.
var linuxNames = []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "microsoft-edge", "brave-browser"}

// candidatePaths builds the ordered list of absolute paths to try, most
// likely first.
func candidatePaths(getenv func(string) string, extraDirs []string) []string {
	var dirs []string
	dirs = append(dirs, extraDirs...)

	var paths []string
	switch runtime.GOOS {
	case "darwin":
		for _, dir := range append(dirs, "/Applications", userDir(getenv, "Applications")) {
			for _, app := range macApps {
				paths = append(paths, filepath.Join(dir, app+".app", "Contents", "MacOS", app))
			}
		}
	case "windows":
		for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "LocalAppData"} {
			if base := getenv(env); base != "" {
				dirs = append(dirs, base)
			}
		}
		for _, dir := range dirs {
			paths = append(paths,
				filepath.Join(dir, "Google", "Chrome", "Application", "chrome.exe"),
				filepath.Join(dir, "Microsoft", "Edge", "Application", "msedge.exe"),
			)
		}
	default: // Linux and other Unix
		dirs = append(dirs, pathDirs(getenv)...)
		dirs = append(dirs, "/usr/bin", "/usr/local/bin")
		for _, dir := range dirs {
			for _, name := range linuxNames {
				paths = append(paths, filepath.Join(dir, name))
			}
		}
	}
	return paths
}

// userDir returns a dir under the user's home directory, "" when HOME is unset.
func userDir(getenv func(string) string, name string) string {
	if home := getenv("HOME"); home != "" {
		return filepath.Join(home, name)
	}
	return ""
}

// pathDirs splits $PATH into its entries, skipping empties.
func pathDirs(getenv func(string) string) []string {
	var dirs []string
	for _, dir := range strings.Split(getenv("PATH"), string(os.PathListSeparator)) {
		if dir != "" {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// executable reports whether path exists as an executable file. On Windows
// any regular file is accepted as executable.
func executable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return runtime.GOOS == "windows" || info.Mode().Perm()&0o111 != 0
}
