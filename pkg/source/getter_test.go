package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadThroughGetter(t *testing.T) {
	tmpDir, _, err := ReadThroughGetter("git::https://github.com/kcl-lang/flask-demo-kcl-manifests.git")
	if err != nil {
		t.Errorf("TestReadThroughGetter() error = %v", err)
		return
	}

	// Check if tmpDir exists
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("TestReadThroughGetter() tmpDir does not exist")
	}

	// Check if kcl.mod file exists in tmpDir
	kclModPath := filepath.Join(tmpDir, "kcl.mod")
	if _, err := os.Stat(kclModPath); os.IsNotExist(err) {
		t.Errorf("TestReadThroughGetter() kcl.mod file does not exist in tmpDir")
	}
}
