package configuration

import (
	"fmt"
	"strings"
)

type urlSource interface {
	provider() string
	serverURL(c *Config) string
	providerURLPrefixes(c *Config, ssh bool) ([]string, error)
	organizationURLPrefixes(c *Config, ssh bool, org string) ([]string, error)
	repositoryURLs(c *Config, ssh bool, repository string) ([]string, error)
}

var urlSources = []urlSource{
	&githubURLSource{},
	&gitlabURLSource{},
	&bitbucketURLSource{},
	&customURLSource{},
}

const (
	GitHubProvider    = "github"
	GitLabProvider    = "gitlab"
	BitbucketProvider = "bitbucket"
	CustomProvider    = "custom"
)

func (c *Config) lookupURLSource() (urlSource, bool) {
	p := c.Provider
	for _, s := range urlSources {
		if p == s.provider() {
			return s, true
		}
	}
	return nil, false

}

func (c *Config) serverURL() string {
	if s, ok := c.lookupURLSource(); ok {
		return s.serverURL(c)
	}
	return ""
}

func (c *Config) allURLs(ssh bool, repository string) ([]string, error) {
	if "*/*" == repository {
		if prefixes, err := c.providerURLPrefixes(ssh); err == nil {
			return prefixes, nil
		} else {
			return nil, err
		}
	}

	if strings.HasSuffix(repository, "/*") {
		if prefixes, err := c.organizationURLPrefixes(ssh, strings.TrimSuffix(repository, "/*")); err == nil {
			return prefixes, nil
		} else {
			return nil, err
		}
	}

	if prefixes, err := c.repositoryURLs(ssh, repository); err == nil {
		return prefixes, nil
	} else {
		return nil, err
	}
}

// providerURLPrefixes returns the URLs used to clone any repository on the provider
func (c *Config) providerURLPrefixes(ssh bool) ([]string, error) {
	if s, ok := c.lookupURLSource(); ok {
		return s.providerURLPrefixes(c, false)
	}
	return nil, fmt.Errorf("unknown/unsupported SCM Provider: %s", c.Provider)
}

// organizationURLPrefixes returns the URLs used to clone any repository in the organization
func (c *Config) organizationURLPrefixes(ssh bool, organization string) ([]string, error) {
	if s, ok := c.lookupURLSource(); ok {
		return s.organizationURLPrefixes(c, false, organization)
	}
	return nil, fmt.Errorf("unknown/unsupported SCM Provider: %s", c.Provider)
}

// repositoryURLs returns the URLs used to clone the repository
func (c *Config) repositoryURLs(ssh bool, repository string) ([]string, error) {
	if s, ok := c.lookupURLSource(); ok {
		return s.repositoryURLs(c, false, repository)
	}
	return nil, fmt.Errorf("unknown/unsupported SCM Provider: %s", c.Provider)
}
