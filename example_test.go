package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kcl-lang.io/krm-kcl/pkg/options"

	pkg "kcl-lang.io/kpm/pkg/package"
)

type fields struct {
	InputPath  string
	OutputPath string
}

type suite struct {
	name    string
	fields  fields
	wantErr bool
}

func TestRunExamples(t *testing.T) {
	var tests []suite
	err := filepath.Walk("./examples", func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(path, "kcl.mod") {
			return nil
		}
		dir := filepath.Dir(path)

		kPkg, err := pkg.LoadKclPkg(dir)
		if err != nil {
			return err
		}
		suiteDir := filepath.Join(dir, "suite")
		goodSuite := filepath.Join(suiteDir, "good.yaml")
		badSuite := filepath.Join(suiteDir, "bad.yaml")
		pkgName := kPkg.GetPkgName()
		if filepath.Base(dir) != pkgName {
			return fmt.Errorf("Mismatch package name %s and base dir %s", pkgName, dir)
		}
		tests = append(tests, suite{
			kPkg.GetPkgName() + "-good-suite",
			fields{
				InputPath: goodSuite,
			},
			false,
		})
		// Bad test suite is optional
		if FileExists(badSuite) {
			tests = append(tests, suite{
				dir + "-bad-suite",
				fields{
					InputPath: badSuite,
				},
				true,
			})
		}
		return nil
	})
	if err != nil {
		t.Errorf("TestRunExamples() error = %v", err)
	}
	fmt.Printf("%d total suites checked\n", len(tests))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &options.RunOptions{
				InputPath:  tt.fields.InputPath,
				OutputPath: tt.fields.OutputPath,
			}
			if err := o.Run(); (err != nil) != tt.wantErr {
				t.Errorf("TestRunExamples() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// FileExists mark whether the path exists.
func FileExists(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil || fi.IsDir() {
		return false
	}
	return true
}
