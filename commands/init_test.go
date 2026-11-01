package commands

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/txtcv/txtcv/schema"
)

func decode(t *testing.T, content []byte, format string) any {
	t.Helper()
	var instance any
	var err error
	switch format {
	case "json":
		err = json.Unmarshal(content, &instance)
	case "yaml":
		err = yaml.Unmarshal(content, &instance)
	default:
		t.Fatalf("unknown format %q", format)
	}
	if err != nil {
		t.Fatalf("decode %s seed: %v", format, err)
	}
	return instance
}

// The two seed files must carry identical CV content; they differ only in
// syntax (and the YAML-only guidance comments).
func TestSeedParity(t *testing.T) {
	y, err := seed("yaml")
	if err != nil {
		t.Fatal(err)
	}
	j, err := seed("json")
	if err != nil {
		t.Fatal(err)
	}
	yamlCV := decode(t, y, "yaml")
	jsonCV := decode(t, j, "json")
	if !reflect.DeepEqual(yamlCV, jsonCV) {
		t.Errorf("seed/cv.yaml and seed/cv.json decode to different content")
	}
}

// Whatever init writes must conform to the JSON Resume schema, as decoded
// exactly as a real consumer would decode it.
func TestSeedValidates(t *testing.T) {
	for _, format := range []string{"yaml", "json"} {
		t.Run(format, func(t *testing.T) {
			content, err := seed(format)
			if err != nil {
				t.Fatal(err)
			}
			if err := schema.Valid(decode(t, content, format)); err != nil {
				t.Error(err)
			}
		})
	}
}

func runInit(t *testing.T, args ...string) error {
	t.Helper()
	full := append([]string{"init"}, args...)
	return initCommand().Run(context.Background(), full)
}

func TestInitWritesDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := runInit(t); err != nil {
		t.Fatalf("init: %v", err)
	}

	got, err := os.ReadFile("cv.yaml")
	if err != nil {
		t.Fatalf("read cv.yaml: %v", err)
	}
	want, _ := seed("yaml")
	if string(got) != string(want) {
		t.Errorf("cv.yaml does not match the YAML seed template")
	}
}

func TestInitFormatFlag(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := runInit(t, "--format", "json"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := os.Stat("cv.json"); err != nil {
		t.Fatalf("cv.json: %v", err)
	}
	if _, err := os.Stat("cv.yaml"); err == nil {
		t.Errorf("cv.yaml should not exist when --format json was given")
	}
}

func TestInitRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := runInit(t); err != nil {
		t.Fatalf("first init: %v", err)
	}
	before, _ := os.ReadFile("cv.yaml")
	if err := runInit(t); err == nil {
		t.Fatal("second init should fail when cv.yaml exists")
	}
	after, _ := os.ReadFile("cv.yaml")
	if string(before) != string(after) {
		t.Errorf("existing cv.yaml was modified by a refused init")
	}
}

func TestInitRejectsBadFormat(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := runInit(t, "--format", "toml"); err == nil {
		t.Fatal("init --format toml should fail")
	}
	if _, err := os.Stat("cv.toml"); err == nil {
		t.Errorf("cv.toml should not exist after a rejected --format")
	}
}
