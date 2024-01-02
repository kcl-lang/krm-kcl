package source

import (
	"fmt"
	"strings"
)

const (
	// GitScheme is the URL scheme used for Git-based requests.
	GitScheme = "git"
	// GitHubDomain is the domain used for GitHub-based requests.
	GitHubDomain = "github.com"
	// GitLabDomain is the domain used for GitLab-based requests.
	GitLabDomain = "gitlab.com/"
	// BitBucketDomain is the domain used for BitBucket-based requests.
	BitBucketDomain = "bitbucket.org/"
)

// IsGit determines whether or not a source is to be treated as a git source.
func IsGit(src string) bool {
	return strings.HasPrefix(src, fmt.Sprintf("%s://", GitScheme))
}

// IsVCSDomain determines whether or not a source is to be treated as a VCS source.
func IsVCSDomain(src string) bool {
	return strings.HasPrefix(src, GitHubDomain) || strings.HasPrefix(src, GitLabDomain) || strings.HasPrefix(src, BitBucketDomain)
}
