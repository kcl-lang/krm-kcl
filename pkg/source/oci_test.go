package source

import "testing"

func TestParseOCIReference(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		wantURL  string
		wantTag  string
	}{
		{
			name:    "no tag",
			ref:     "ghcr.io/kcl-lang/helloworld",
			wantURL: "oci://ghcr.io/kcl-lang/helloworld",
			wantTag: "",
		},
		{
			name:    "path tag",
			ref:     "ghcr.io/kcl-lang/set-annotations:0.1.1",
			wantURL: "oci://ghcr.io/kcl-lang/set-annotations",
			wantTag: "0.1.1",
		},
		{
			name:    "path tag with multiple slashes",
			ref:     "registry.example.com/team/sub/project:v1.2.3",
			wantURL: "oci://registry.example.com/team/sub/project",
			wantTag: "v1.2.3",
		},
		{
			name:    "port without tag",
			ref:     "host.docker.internal:7900/myapp",
			wantURL: "oci://host.docker.internal:7900/myapp",
			wantTag: "",
		},
		{
			name:    "port with query tag",
			ref:     "host.docker.internal:7900/myapp?tag=0.0.1",
			wantURL: "oci://host.docker.internal:7900/myapp",
			wantTag: "0.0.1",
		},
		{
			name:    "port with path tag",
			ref:     "host.docker.internal:7900/myapp:v1",
			wantURL: "oci://host.docker.internal:7900/myapp",
			wantTag: "v1",
		},
		{
			name:    "query tag wins over path tag",
			ref:     "host:7900/myapp:v1?tag=0.0.1",
			wantURL: "oci://host:7900/myapp",
			wantTag: "0.0.1",
		},
		{
			name:    "empty tag value is ignored",
			ref:     "ghcr.io/kcl-lang/helloworld?tag=",
			wantURL: "oci://ghcr.io/kcl-lang/helloworld",
			wantTag: "",
		},
		{
			name:    "query with non-tag parameters is left alone",
			ref:     "ghcr.io/kcl-lang/helloworld?other=value",
			wantURL: "oci://ghcr.io/kcl-lang/helloworld",
			wantTag: "",
		},
		{
			name:    "port 443 with query tag",
			ref:     "myregistry.example.com:443/kcl/foo?tag=1.2.3",
			wantURL: "oci://myregistry.example.com:443/kcl/foo",
			wantTag: "1.2.3",
		},
		{
			name:    "host without path has no tag",
			ref:     "host:7900",
			wantURL: "oci://host:7900",
			wantTag: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotTag := ParseOCIReference(tt.ref)
			if gotURL != tt.wantURL {
				t.Errorf("ParseOCIReference(%q) url = %q, want %q", tt.ref, gotURL, tt.wantURL)
			}
			if gotTag != tt.wantTag {
				t.Errorf("ParseOCIReference(%q) tag = %q, want %q", tt.ref, gotTag, tt.wantTag)
			}
		})
	}
}