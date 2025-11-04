package configuration

import (
	"net/url"
	"strings"
)

type bitbucketDatacenterURLSource struct{}

func (s bitbucketDatacenterURLSource) provider() string {
	return BitbucketDatacenterProvider
}

func (s bitbucketDatacenterURLSource) serverURL(c *Config) string {
	return c.BitbucketServerURL
}

func (s bitbucketDatacenterURLSource) providerURLPrefixes(c *Config, ssh bool) ([]string, error) {
	parsed, err := url.Parse(c.BitbucketServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath("/")
	if !ssh {
		return []string{parsed.String()}, nil
	}
	// TODO for Bitbucket server, query the API and discover the ssh port
	return []string{
		"git@" + preferred.Hostname() + ":" + strings.TrimPrefix(preferred.Path, "/"),
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
	}, nil
}

func (s bitbucketDatacenterURLSource) organizationURLPrefixes(c *Config, ssh bool, org string) ([]string, error) {
	parsed, err := url.Parse(c.BitbucketServerURL)
	if err != nil {
		return nil, err
	}
	preferred := parsed.JoinPath(org + "/")
	if !ssh {
		return []string{parsed.String()}, nil
	}
	// TODO for Bitbucket server, query the API and discover the ssh port
	return []string{
		"git@" + preferred.Hostname() + ":" + strings.TrimPrefix(preferred.Path, "/"),
		"ssh://git@" + preferred.Hostname() + "/" + strings.TrimPrefix(preferred.Path, "/"),
	}, nil
}

func (s bitbucketDatacenterURLSource) repositoryURLs(c *Config, ssh bool, repository string) ([]string, error) {
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
