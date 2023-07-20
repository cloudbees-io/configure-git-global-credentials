package configuration

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	GitHubProvider    = "github"
	GitLabProvider    = "gitlab"
	BitbucketProvider = "bitbucket"
	CustomProvider    = "custom"
)

func (c *Config) serverURL() string {
	p := c.Provider
	switch p {
	case GitHubProvider:
		return c.GitHubServerURL
	case BitbucketProvider:
		return c.BitbucketServerURL
	case GitLabProvider:
		return c.GitLabServerURL
	default:
		return ""
	}
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
	p := c.Provider
	switch p {
	case GitHubProvider:
		return c.githubProviderURLPrefixes(ssh)
	case BitbucketProvider:
		return c.bitbucketProviderURLPrefixes(ssh)
	case GitLabProvider:
		return c.gitlabProviderURLPrefixes(ssh)
	case CustomProvider:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown/unsupported SCM Provider: %s", p)
	}
}

// organizationURLPrefixes returns the URLs used to clone any repository in the organization
func (c *Config) organizationURLPrefixes(ssh bool, organization string) ([]string, error) {
	p := c.Provider
	switch p {
	case GitHubProvider:
		return c.githubOrganizationURLPrefixes(ssh, organization)
	case BitbucketProvider:
		return c.bitbucketOrganizationURLPrefixes(ssh, organization)
	case GitLabProvider:
		return c.gitlabOrganizationURLPrefixes(ssh, organization)
	case CustomProvider:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown/unsupported SCM Provider: %s", p)
	}
}

// repositoryURLs returns the URLs used to clone the repository
func (c *Config) repositoryURLs(ssh bool, repository string) ([]string, error) {
	p := c.Provider
	switch p {
	case GitHubProvider:
		return c.githubRepositoryURLs(ssh, repository)
	case BitbucketProvider:
		return c.bitbucketRepositoryURLs(ssh, repository)
	case GitLabProvider:
		return c.gitlabRepositoryURLs(ssh, repository)
	case CustomProvider:
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
	default:
		return nil, fmt.Errorf("unknown/unsupported SCM Provider: %s", p)
	}
}

func (c *Config) githubRepositoryURLs(ssh bool, repository string) ([]string, error) {
	parsed, err := url.Parse(c.GitHubServerURL)
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

func (c *Config) bitbucketRepositoryURLs(ssh bool, repository string) ([]string, error) {
	parsed, err := url.Parse(c.BitbucketServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath(repository + ".git")
	if !ssh {
		return []string{preferred.String()}, nil
	}
	// TODO for Bitbucket server, query the API and discover the ssh port
	return []string{
		"git@" + preferred.Hostname() + ":" + preferred.Path,
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
	}, nil
}

func (c *Config) gitlabRepositoryURLs(ssh bool, repository string) ([]string, error) {
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

func (c *Config) githubOrganizationURLPrefixes(ssh bool, organization string) ([]string, error) {
	parsed, err := url.Parse(c.GitHubServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath(organization + "/")
	if !ssh {
		return []string{preferred.String()}, nil
	}
	return []string{
		"git@" + preferred.Hostname() + ":" + strings.TrimPrefix(preferred.Path, "/"),
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
	}, nil
}

func (c *Config) bitbucketOrganizationURLPrefixes(ssh bool, organization string) ([]string, error) {
	parsed, err := url.Parse(c.BitbucketServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath(organization + "/")
	if !ssh {
		return []string{preferred.String()}, nil
	}
	// TODO for Bitbucket server, query the API and discover the ssh port
	return []string{
		"git@" + preferred.Hostname() + ":" + strings.TrimPrefix(preferred.Path, "/"),
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
	}, nil
}

func (c *Config) gitlabOrganizationURLPrefixes(ssh bool, organization string) ([]string, error) {
	parsed, err := url.Parse(c.GitLabServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath(organization + "/")
	if !ssh {
		return []string{preferred.String()}, nil
	}
	return []string{
		"git@" + preferred.Hostname() + ":" + strings.TrimPrefix(preferred.Path, "/"),
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
	}, nil
}

func (c *Config) githubProviderURLPrefixes(ssh bool) ([]string, error) {
	parsed, err := url.Parse(c.GitHubServerURL)
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

func (c *Config) bitbucketProviderURLPrefixes(ssh bool) ([]string, error) {
	parsed, err := url.Parse(c.BitbucketServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath("/")
	if !ssh {
		return []string{preferred.String()}, nil
	}
	// TODO for Bitbucket server, query the API and discover the ssh port
	return []string{
		"git@" + preferred.Hostname() + ":" + strings.TrimPrefix(preferred.Path, "/"),
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
	}, nil
}

func (c *Config) gitlabProviderURLPrefixes(ssh bool) ([]string, error) {
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
