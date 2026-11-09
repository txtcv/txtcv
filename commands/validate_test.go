package commands

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name    string
		files   map[string]string
		args    []string
		want    string
		wantErr string
	}{
		{
			name:  "default YAML",
			files: map[string]string{"cv.yaml": "basics:\n  name: Ada\n"},
			want:  "cv.yaml is valid JSON Resume.\n",
		},
		{
			name:  "default JSON",
			files: map[string]string{"cv.json": `{"basics":{"name":"Ada"}}`},
			want:  "cv.json is valid JSON Resume.\n",
		},
		{
			name:  "prefer YAML",
			files: map[string]string{"cv.yaml": "basics: {}\n", "cv.json": "invalid"},
			want:  "cv.yaml is valid JSON Resume.\n",
		},
		{
			name:  "explicit file overrides default",
			files: map[string]string{"resume.yml": "basics: {}\n", "cv.yaml": "invalid"},
			args:  []string{"resume.yml"},
			want:  "resume.yml is valid JSON Resume.\n",
		},
		{
			name:  "partial dates and vendor metadata",
			files: map[string]string{"resume.json": `{"work":[{"startDate":"2024","endDate":"2024-06"},{"startDate":"2024-06-15"}],"meta":{"txtcv":{"theme":"code"}}}`},
			args:  []string{"resume.json"},
			want:  "resume.json is valid JSON Resume.\n",
		},
		{name: "missing default", wantErr: "run `txtcv init`"},
		{name: "missing explicit file", args: []string{"missing.json"}, wantErr: "read missing.json:"},
		{name: "extra arguments", args: []string{"a.json", "b.json"}, wantErr: "usage: txtcv validate [file]"},
		{name: "malformed JSON", files: map[string]string{"cv.json": `{"basics":`}, wantErr: "decode cv.json:"},
		{name: "malformed YAML", files: map[string]string{"cv.yaml": "basics: ["}, wantErr: "parse cv.yaml:"},
		{name: "invalid JSON field", files: map[string]string{"cv.json": `{"basics":{"name":42}}`}, wantErr: "is not valid JSON Resume:"},
		{name: "invalid YAML field", files: map[string]string{"cv.yaml": "work: wrong\n"}, wantErr: "is not valid JSON Resume:"},
		{name: "invalid date", files: map[string]string{"cv.json": `{"work":[{"startDate":"yesterday"}]}`}, wantErr: "is not valid JSON Resume:"},
		{name: "empty document", files: map[string]string{"cv.yaml": ""}, wantErr: "is empty:"},
		{name: "null document", files: map[string]string{"cv.json": "null"}, wantErr: "is empty:"},
		{name: "array document", files: map[string]string{"cv.json": "[]"}, wantErr: "decode cv.json:"},
		{name: "unsupported format", files: map[string]string{"resume.txt": "{}"}, args: []string{"resume.txt"}, wantErr: "unsupported CV file format"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			for path, content := range tc.files {
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			var out bytes.Buffer
			cmd := Root()
			cmd.Writer = &out
			err := cmd.Run(context.Background(), append([]string{"txtcv", "validate"}, tc.args...))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v, want containing %q", err, tc.wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if got := out.String(); got != tc.want {
				t.Errorf("output = %q, want %q", got, tc.want)
			}
			for path, want := range tc.files {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != want {
					t.Errorf("validation modified %s: content = %q, error = %v", path, got, err)
				}
			}
		})
	}
}
