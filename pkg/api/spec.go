package api

const (
	// KCLRunKind represents the kind of resource for the KCLRun resource.
	KCLRunKind = "KCLRun"

	// SourceKey is the key for the source field in the KCLRun resource, which denotes the source of the KCL script
	// and it can be from the inline code, git, oci source, etc.
	SourceKey = "source"

	// ParamsKey is the key for the params field in the KCLRun resource, which denotes the top level dynamic arguments.
	ParamsKey = "params"

	// ConfigKey is the key for the config field in the KCLRun resource, which denotes the KCL CLI running config.
	ConfigKey = "config"

	// MatchConstraintsKey is the key for the match constraints field in the KCLRun resource.
	MatchConstraintsKey = "matchConstraints"
)

// ConfigSpec defines the compile config.
type ConfigSpec struct {
	// Arguments is the list of top level dynamic arguments for the kcl option function, e.g., env="prod"
	Arguments []string `json:"arguments" yaml:"arguments"`
	// Settings is the list of kcl setting files including all of the CLI config.
	Settings []string `json:"settings" yaml:"settings"`
	// Overrides is the list of override paths and values, e.g., app.image="v2"
	Overrides []string `json:"overrides" yaml:"overrides"`
	// PathSelectors is the list of path selectors to select output result, e.g., a.b.c
	PathSelectors []string `json:"pathSelectors" yaml:"pathSelectors"`
	// Vendor denotes running kcl in the vendor mode.
	Vendor bool `json:"vendor" yaml:"vendor"`
	// SortKeys denotes sorting the output result keys, e.g., `{b = 1, a = 2} => {a = 2, b = 1}`.
	SortKeys bool `json:"sortKeys" yaml:"sortKeys"`
	// ShowHidden denotes output the hidden attribute in the result.
	ShowHidden bool `json:"showHidden" yaml:"showHidden"`
	// DisableNone denotes running kcl and disable dumping None values.
	DisableNone bool `json:"disableNone" yaml:"disableNone"`
	// Debug denotes running kcl in debug mode.
	Debug bool `json:"debug" yaml:"debug"`
	// StrictRangeCheck performs the 32-bit strict numeric range checks on numbers.
	StrictRangeCheck bool `json:"strictRangeCheck" yaml:"strictRangeCheck"`
}

// CredSpec defines authentication credentials for remote locations
type CredSpec struct {
	Url      string `json:"url" yaml:"url"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// MatchConstraintsSpec defines the resource matching rules.
type MatchConstraintsSpec struct {
	ResourceRules []ResourceRule `json:"resourceRules,omitempty" yaml:"resourceRules,omitempty"`
}

// ResourceRule defines a rule for matching resources.
type ResourceRule struct {
	APIVersions []string `json:"apiVersions,omitempty" yaml:"apiVersions,omitempty"`
	Kinds       []string `json:"kinds,omitempty" yaml:"kinds,omitempty"`
}
