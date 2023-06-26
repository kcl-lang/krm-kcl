package options

import (
	"testing"
)

func TestPipeline(t *testing.T) {
	o := &RunOptions{
		InputPath: "./testdata/kcl-run.yaml",
	}
	if err := o.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestPipelineLocalPath(t *testing.T) {
	o := &RunOptions{
		InputPath: "./testdata/kcl-run-local.yaml",
	}
	if err := o.Run(); err != nil {
		t.Fatal(err)
	}
}

// TODO: OCI source
// func TestPipelineOCI(t *testing.T) {
// 	o := &RunOptions{
// 		InputPath: "./testdata/kcl-run-oci.yaml",
// 	}
// 	if err := o.Run(); err != nil {
// 		t.Fatal(err)
// 	}
// }

func TestPipelineGit(t *testing.T) {
	o := &RunOptions{
		InputPath: "./testdata/kcl-run-git.yaml",
	}
	if err := o.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestPipelineHttps(t *testing.T) {
	o := &RunOptions{
		InputPath: "./testdata/kcl-run-https.yaml",
	}
	if err := o.Run(); err != nil {
		t.Fatal(err)
	}
}
