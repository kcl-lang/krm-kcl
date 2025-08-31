package options

import (
	"testing"
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

func TestRunCode(t *testing.T) {
	tests := []suite{
		{
			"resource_list",
			fields{
				InputPath: "./testdata/resource_list/kcl-run-code.yaml",
			},
			false,
		},
		{
			"yaml_stream",
			fields{
				InputPath: "./testdata/yaml_stream/kcl-run-code.yaml",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &RunOptions{
				InputPath:  tt.fields.InputPath,
				OutputPath: tt.fields.OutputPath,
			}
			if err := o.Run(); (err != nil) != tt.wantErr {
				t.Errorf("TestRunCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLocalPath(t *testing.T) {
	tests := []suite{
		{
			"resource_list",
			fields{
				InputPath: "./testdata/resource_list/kcl-run-local.yaml",
			},
			false,
		},
		{
			"yaml_stream",
			fields{
				InputPath: "./testdata/yaml_stream/kcl-run-local.yaml",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &RunOptions{
				InputPath:  tt.fields.InputPath,
				OutputPath: tt.fields.OutputPath,
			}
			if err := o.Run(); (err != nil) != tt.wantErr {
				t.Errorf("TestRunLocalPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunOCI(t *testing.T) {
	tests := []suite{
		{
			"resource_list",
			fields{
				InputPath: "./testdata/resource_list/kcl-run-oci.yaml",
			},
			false,
		},
		{
			"resource_list",
			fields{
				InputPath: "./testdata/resource_list/kcl-run-oci-with-version.yaml",
			},
			false,
		},
		{
			"resource_list",
			fields{
				InputPath: "./testdata/resource_list/kcl-run-oci-with-bad-version.yaml",
			},
			true,
		},
		{
			"yaml_stream",
			fields{
				InputPath: "./testdata/yaml_stream/kcl-run-oci.yaml",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &RunOptions{
				InputPath:  tt.fields.InputPath,
				OutputPath: tt.fields.OutputPath,
			}
			if err := o.Run(); (err != nil) != tt.wantErr {
				t.Errorf("TestRunOCI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunGit(t *testing.T) {
	tests := []suite{
		{
			"resource_list",
			fields{
				InputPath: "./testdata/resource_list/kcl-run-git.yaml",
			},
			false,
		},
		{
			"yaml_stream",
			fields{
				InputPath: "./testdata/yaml_stream/kcl-run-git.yaml",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &RunOptions{
				InputPath:  tt.fields.InputPath,
				OutputPath: tt.fields.OutputPath,
			}
			if err := o.Run(); (err != nil) != tt.wantErr {
				t.Errorf("TestRunGit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunHttps(t *testing.T) {
	tests := []suite{
		{
			"resource_list",
			fields{
				InputPath: "./testdata/resource_list/kcl-run-https.yaml",
			},
			false,
		},
		{
			"yaml_stream",
			fields{
				InputPath: "./testdata/yaml_stream/kcl-run-https.yaml",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &RunOptions{
				InputPath:  tt.fields.InputPath,
				OutputPath: tt.fields.OutputPath,
			}
			if err := o.Run(); (err != nil) != tt.wantErr {
				t.Errorf("TestRunHttps() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
