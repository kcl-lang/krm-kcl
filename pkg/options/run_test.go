package options

import (
	"testing"
)

func TestRunCode(t *testing.T) {
	o := &RunOptions{
		InputPath: "./testdata/kcl-run-code.yaml",
	}
	if err := o.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestRunLocalPath(t *testing.T) {
	o := &RunOptions{
		InputPath: "./testdata/kcl-run-local.yaml",
	}
	if err := o.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestRunOCI(t *testing.T) {
	o := &RunOptions{
		InputPath: "./testdata/kcl-run-oci.yaml",
	}
	if err := o.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestRunGit(t *testing.T) {
	o := &RunOptions{
		InputPath: "./testdata/kcl-run-git.yaml",
	}
	if err := o.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestRunHttps(t *testing.T) {
	o := &RunOptions{
		InputPath: "./testdata/kcl-run-https.yaml",
	}
	if err := o.Run(); err != nil {
		t.Fatal(err)
	}
}
