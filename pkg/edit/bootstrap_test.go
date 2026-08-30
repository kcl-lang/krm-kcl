package edit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSourceToTempEntryInlineCode(t *testing.T) {
	const code = `a = 1`
	entry, err := SourceToTempEntry(code)
	if err != nil {
		t.Fatalf("SourceToTempEntry() error = %v", err)
	}
	defer KCLEntryOriginTmpDirCleanup(entry)

	if entry.tmpDir == "" {
		t.Fatal("inline code should be materialized into a temp dir, got empty tmpDir")
	}
	if filepath.Base(entry.source) != "prog.k" {
		t.Errorf("inline code should be written to prog.k, got %q", entry.source)
	}
	content, err := os.ReadFile(entry.source)
	if err != nil {
		t.Fatalf("reading generated entry file: %v", err)
	}
	if string(content) != code {
		t.Errorf("entry file content = %q, want %q", string(content), code)
	}
}

func TestSourceToTempEntryPassthroughSources(t *testing.T) {
	abs, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("resolving abs path: %v", err)
	}
	tests := []struct {
		name string
		src  string
	}{
		{"oci", "oci://ghcr.io/kcl-lang/set-annotations"},
		{"oci_with_tag", "oci://ghcr.io/kcl-lang/set-annotations:0.1.0"},
		{"relative_local", "./main.k"},
		{"absolute_local", abs},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := SourceToTempEntry(tt.src)
			if err != nil {
				t.Fatalf("SourceToTempEntry() error = %v", err)
			}
			defer KCLEntryOriginTmpDirCleanup(entry)

			// OCI and local sources are consumed in place, so they must not be
			// copied into a temp dir nor rewritten.
			if entry.source != tt.src {
				t.Errorf("source = %q, want it passed through unchanged as %q", entry.source, tt.src)
			}
			if entry.tmpDir != "" {
				t.Errorf("tmpDir = %q, want empty (no temp dir for in-place sources)", entry.tmpDir)
			}
		})
	}
}

func TestKCLEntryOriginTmpDirCleanup(t *testing.T) {
	entry, err := SourceToTempEntry(`a = 1`)
	if err != nil {
		t.Fatalf("SourceToTempEntry() error = %v", err)
	}
	tmpDir := entry.tmpDir

	KCLEntryOriginTmpDirCleanup(entry)
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Errorf("temp dir %q still exists after cleanup", tmpDir)
	}
	// Cleanup must stay safe when the dir is already gone or was never created.
	KCLEntryOriginTmpDirCleanup(entry)
	KCLEntryOriginTmpDirCleanup(&KCLEntryOrigin{source: "oci://example.com/pkg"})
}

func TestToKCLValueString(t *testing.T) {
	node, err := yaml.Parse(`
replicas: 2
name: nginx
ports:
- 80
- 443
`)
	if err != nil {
		t.Fatalf("parsing yaml: %v", err)
	}
	got, err := ToKCLValueString(node, emptyConfig)
	if err != nil {
		t.Fatalf("ToKCLValueString() error = %v", err)
	}
	for _, want := range []string{`"replicas":2`, `"name":"nginx"`, `"ports":[80,443]`} {
		if !strings.Contains(got, want) {
			t.Errorf("ToKCLValueString() = %s, want it to contain %s", got, want)
		}
	}
}

func TestToKCLValueStringNilFallsBackToDefault(t *testing.T) {
	node, err := yaml.Parse(`items: []`)
	if err != nil {
		t.Fatalf("parsing yaml: %v", err)
	}
	// Looking up a missing field yields a nil RNode, which must produce the default.
	missing, err := node.Pipe(yaml.Lookup("functionConfig", "spec", "params"))
	if err != nil {
		t.Fatalf("lookup error: %v", err)
	}
	got, err := ToKCLValueString(missing, emptyConfig)
	if err != nil {
		t.Fatalf("ToKCLValueString() error = %v", err)
	}
	if got != emptyConfig {
		t.Errorf("ToKCLValueString() = %q, want default %q", got, emptyConfig)
	}
}
