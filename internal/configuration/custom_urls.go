package configuration

import (
	"fmt"
	"net/url"
)

type customURLSource struct{}

func (s customURLSource) provider() string {
	return CustomProvider
}

func (s customURLSource) serverURL(_ *Config) string {
	return ""
}

func (s customURLSource) providerURLPrefixes(c *Config, ssh bool) ([]string, error) {
	return nil, nil
}

func (s customURLSource) organizationURLPrefixes(c *Config, ssh bool, org string) ([]string, error) {
	return nil, nil
}

func (s customURLSource) repositoryURLs(c *Config, ssh bool, repository string) ([]string, error) {
	if u, err := url.Parse(repository); err == nil {
		if !ssh == (u.Scheme == "http" || u.Scheme == "https") {
			target := "http(s)"
			if ssh {
				target = "ssh"
			}
			return nil, fmt.Errorf("cannot convert custom provider clone url %s into %s form", repository, target)
		}
		return []string{repository}, nil
	} else {
		return nil, fmt.Errorf("could not parse repository url %s: %w", repository, err)
	}
}
