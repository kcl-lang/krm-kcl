package config

import (
	"strings"

	"kcl-lang.io/krm-kcl/pkg/api"
	"kcl-lang.io/krm-kcl/pkg/kube"
)

// MatchResourceRules checks if the given Kubernetes object matches the resource rules specified in KCLRun.
func MatchResourceRules(obj *kube.KubeObject, MatchConstraints *api.MatchConstraintsSpec) bool {
	// if MatchConstraints.ResourceRules is not set (nil or empty), return true by default
	if len(MatchConstraints.ResourceRules) == 0 {
		return true
	}
	// iterate through each resource rule
	for _, rule := range MatchConstraints.ResourceRules {
		if containsString(rule.APIVersions, obj.GetAPIVersion()) &&
			containsString(rule.Kinds, obj.GetKind()) {
			return true
		}
	}
	// if no match is found, return false
	return false
}

// isOk checks if a given string is in the list of "OK" values.
func isOk(value string) bool {
	okValues := []string{"ok", "yes", "true", "1", "on"}
	for _, v := range okValues {
		if strings.EqualFold(strings.ToLower(value), strings.ToLower(v)) {
			return true
		}
	}
	return false
}

// containsString checks if a slice contains a string or "*"
func containsString(slice []string, str string) bool {
	if len(slice) == 0 {
		return true
	}
	for _, s := range slice {
		if s == "*" || s == str {
			return true
		}
	}
	return false
}
