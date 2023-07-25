package configuration

import (
	"net/url"
	"strings"
)

type gitlabURLSource struct{}

func (s gitlabURLSource) provider() string {
	return GitLabProvider
}

func (s gitlabURLSource) serverURL(c *Config) string {
	return c.GitLabServerURL
}

func (s gitlabURLSource) providerURLPrefixes(c *Config, ssh bool) ([]string, error) {
	parsed, err := url.Parse(c.GitLabServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath("/")
	if !ssh {
		return []string{preferred.String()}, nil
	}
	return []string{
		"git@" + preferred.Hostname() + ":" + strings.TrimPrefix(preferred.Path, "/"),
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
	}, nil
}

func (s gitlabURLSource) organizationURLPrefixes(c *Config, ssh bool, org string) ([]string, error) {
	parsed, err := url.Parse(c.GitLabServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath(org + "/")
	if !ssh {
		return []string{preferred.String()}, nil
	}
	return []string{
		"git@" + preferred.Hostname() + ":" + strings.TrimPrefix(preferred.Path, "/"),
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
	}, nil
}

func (s gitlabURLSource) repositoryURLs(c *Config, ssh bool, repository string) ([]string, error) {
	parsed, err := url.Parse(c.GitLabServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath(repository + ".git")
	accepted := parsed.JoinPath(repository)
	if !ssh {
		return []string{preferred.String(), accepted.String()}, nil
	}
	return []string{
		"git@" + preferred.Hostname() + ":" + strings.TrimPrefix(preferred.Path, "/"),
		"git@" + accepted.Hostname() + ":" + strings.TrimPrefix(accepted.Path, "/"),
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
		"ssh://git@" + accepted.Hostname() + "/" + strings.TrimPrefix(accepted.Path, "/"),
	}, nil
}
