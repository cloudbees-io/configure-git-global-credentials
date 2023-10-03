package configuration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_insteadOfURLs(t *testing.T) {
	type fields struct {
		Provider           string
		Repositories       string
		SshKey             string
		GitHubServerURL    string
		BitbucketServerURL string
		GitLabServerURL    string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string][]string
	}{
		{
			name: "github cloud everything to https",
			fields: fields{
				Provider:        "github",
				Repositories:    "*/*",
				SshKey:          "",
				GitHubServerURL: "https://github.com",
			},
			want: map[string][]string{
				"https://github.com": {
					"git@github.com:",
					"ssh://git@github.com/",
				},
			},
		},
		{
			name: "github cloud everything to ssh",
			fields: fields{
				Provider:        "github",
				Repositories:    "*/*",
				SshKey:          "-dummy-",
				GitHubServerURL: "https://github.com",
			},
			want: map[string][]string{
				"git@github.com:": {
					"ssh://git@github.com/",
					"https://github.com/",
				},
			},
		},
		{
			name: "github cloud org to https",
			fields: fields{
				Provider:        "github",
				Repositories:    "example/*",
				SshKey:          "",
				GitHubServerURL: "https://github.com",
			},
			want: map[string][]string{
				"https://github.com/example/": {
					"git@github.com:example/",
					"ssh://git@github.com/example/",
				},
			},
		},
		{
			name: "github cloud org to ssh",
			fields: fields{
				Provider:           "github",
				Repositories:       "example/*",
				SshKey:             "-dummy-",
				GitHubServerURL:    "https://github.com",
				BitbucketServerURL: "",
				GitLabServerURL:    "",
			},
			want: map[string][]string{
				"git@github.com:example/": {
					"ssh://git@github.com/example/",
					"https://github.com/example/",
				},
			},
		},
		{
			name: "github cloud repo to https",
			fields: fields{
				Provider:        "github",
				Repositories:    "example/foo",
				SshKey:          "",
				GitHubServerURL: "https://github.com",
			},
			want: map[string][]string{
				"https://github.com/example/foo.git": {
					"https://github.com/example/foo",
					"git@github.com:example/foo.git",
					"git@github.com:example/foo",
					"ssh://git@github.com/example/foo.git",
					"ssh://git@github.com/example/foo",
				},
			},
		},
		{
			name: "github cloud some repos to ssh",
			fields: fields{
				Provider:        "github",
				Repositories:    "example/foo",
				SshKey:          "-dummy-",
				GitHubServerURL: "https://github.com",
			},
			want: map[string][]string{
				"git@github.com:example/foo.git": {
					"git@github.com:example/foo",
					"ssh://git@github.com/example/foo.git",
					"ssh://git@github.com/example/foo",
					"https://github.com/example/foo.git",
					"https://github.com/example/foo",
				},
			},
		},
		{
			name: "github enterprise everything to https",
			fields: fields{
				Provider:        "github",
				Repositories:    "*/*",
				SshKey:          "",
				GitHubServerURL: "https://my-ghe.example.com",
			},
			want: map[string][]string{
				"https://my-ghe.example.com/": {
					"git@my-ghe.example.com:",
					"ssh://git@my-ghe.example.com/",
				},
			},
		},
		{
			name: "github enterprise everything to ssh",
			fields: fields{
				Provider:        "github",
				Repositories:    "*/*",
				SshKey:          "-dummy-",
				GitHubServerURL: "https://my-ghe.example.com",
			},
			want: map[string][]string{
				"git@my-ghe.example.com:": {
					"ssh://git@my-ghe.example.com/",
					"https://my-ghe.example.com/",
				},
			},
		},
		{
			name: "github enterprise org to https",
			fields: fields{
				Provider:        "github",
				Repositories:    "example/*",
				SshKey:          "",
				GitHubServerURL: "https://my-ghe.example.com",
			},
			want: map[string][]string{
				"https://my-ghe.example.com/example/": {
					"git@my-ghe.example.com:example/",
					"ssh://git@my-ghe.example.com/example/",
				},
			},
		},
		{
			name: "github enterprise org to ssh",
			fields: fields{
				Provider:        "github",
				Repositories:    "example/*",
				SshKey:          "-dummy-",
				GitHubServerURL: "https://my-ghe.example.com",
			},
			want: map[string][]string{
				"git@my-ghe.example.com:example/": {
					"ssh://git@my-ghe.example.com/example/",
					"https://my-ghe.example.com/example/",
				},
			},
		},
		{
			name: "github enterprise repo to https",
			fields: fields{
				Provider:        "github",
				Repositories:    "example/foo",
				SshKey:          "",
				GitHubServerURL: "https://my-ghe.example.com",
			},
			want: map[string][]string{
				"https://my-ghe.example.com/example/foo.git": {
					"https://my-ghe.example.com/example/foo",
					"git@my-ghe.example.com:example/foo.git",
					"git@my-ghe.example.com:example/foo",
					"ssh://git@my-ghe.example.com/example/foo.git",
					"ssh://git@my-ghe.example.com/example/foo",
				},
			},
		},
		{
			name: "github enterprise some repos to ssh",
			fields: fields{
				Provider:        "github",
				Repositories:    "example/foo",
				SshKey:          "-dummy-",
				GitHubServerURL: "https://my-ghe.example.com",
			},
			want: map[string][]string{
				"git@my-ghe.example.com:example/foo.git": {
					"git@my-ghe.example.com:example/foo",
					"ssh://git@my-ghe.example.com/example/foo.git",
					"ssh://git@my-ghe.example.com/example/foo",
					"https://my-ghe.example.com/example/foo.git",
					"https://my-ghe.example.com/example/foo",
				},
			},
		},
		{
			name: "gitlab cloud everything to https",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "*/*",
				SshKey:          "",
				GitLabServerURL: "https://gitlab.com",
			},
			want: map[string][]string{
				"https://gitlab.com/": {
					"git@gitlab.com:",
					"ssh://git@gitlab.com/",
				},
			},
		},
		{
			name: "gitlab cloud everything to ssh",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "*/*",
				SshKey:          "-dummy-",
				GitLabServerURL: "https://gitlab.com",
			},
			want: map[string][]string{
				"git@gitlab.com:": {
					"ssh://git@gitlab.com/",
					"https://gitlab.com/",
				},
			},
		},
		{
			name: "gitlab cloud org to https",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "example/*",
				SshKey:          "",
				GitLabServerURL: "https://gitlab.com",
			},
			want: map[string][]string{
				"https://gitlab.com/example/": {
					"git@gitlab.com:example/",
					"ssh://git@gitlab.com/example/",
				},
			},
		},
		{
			name: "gitlab cloud org to ssh",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "example/*",
				SshKey:          "-dummy-",
				GitLabServerURL: "https://gitlab.com",
			},
			want: map[string][]string{
				"git@gitlab.com:example/": {
					"ssh://git@gitlab.com/example/",
					"https://gitlab.com/example/",
				},
			},
		},
		{
			name: "gitlab cloud repo to https",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "example/foo",
				SshKey:          "",
				GitLabServerURL: "https://gitlab.com",
			},
			want: map[string][]string{
				"https://gitlab.com/example/foo.git": {
					"https://gitlab.com/example/foo",
					"git@gitlab.com:example/foo.git",
					"git@gitlab.com:example/foo",
					"ssh://git@gitlab.com/example/foo.git",
					"ssh://git@gitlab.com/example/foo",
				},
			},
		},
		{
			name: "gitlab cloud some repos to ssh",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "example/foo",
				SshKey:          "-dummy-",
				GitLabServerURL: "https://gitlab.com",
			},
			want: map[string][]string{
				"git@gitlab.com:example/foo.git": {
					"git@gitlab.com:example/foo",
					"ssh://git@gitlab.com/example/foo.git",
					"ssh://git@gitlab.com/example/foo",
					"https://gitlab.com/example/foo.git",
					"https://gitlab.com/example/foo",
				},
			},
		},
		{
			name: "gitlab enterprise everything to https",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "*/*",
				SshKey:          "",
				GitLabServerURL: "https://my-gle.example.com",
			},
			want: map[string][]string{
				"https://my-gle.example.com/": {
					"git@my-gle.example.com:",
					"ssh://git@my-gle.example.com/",
				},
			},
		},
		{
			name: "gitlab enterprise everything to ssh",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "*/*",
				SshKey:          "-dummy-",
				GitLabServerURL: "https://my-gle.example.com",
			},
			want: map[string][]string{
				"git@my-gle.example.com:": {
					"ssh://git@my-gle.example.com/",
					"https://my-gle.example.com/",
				},
			},
		},
		{
			name: "gitlab enterprise org to https",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "example/*",
				SshKey:          "",
				GitLabServerURL: "https://my-gle.example.com",
			},
			want: map[string][]string{
				"https://my-gle.example.com/example/": {
					"git@my-gle.example.com:example/",
					"ssh://git@my-gle.example.com/example/",
				},
			},
		},
		{
			name: "gitlab enterprise org to ssh",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "example/*",
				SshKey:          "-dummy-",
				GitLabServerURL: "https://my-gle.example.com",
			},
			want: map[string][]string{
				"git@my-gle.example.com:example/": {
					"ssh://git@my-gle.example.com/example/",
					"https://my-gle.example.com/example/",
				},
			},
		},
		{
			name: "gitlab enterprise repo to https",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "example/foo",
				SshKey:          "",
				GitLabServerURL: "https://my-gle.example.com",
			},
			want: map[string][]string{
				"https://my-gle.example.com/example/foo.git": {
					"https://my-gle.example.com/example/foo",
					"git@my-gle.example.com:example/foo.git",
					"git@my-gle.example.com:example/foo",
					"ssh://git@my-gle.example.com/example/foo.git",
					"ssh://git@my-gle.example.com/example/foo",
				},
			},
		},
		{
			name: "gitlab enterprise some repos to ssh",
			fields: fields{
				Provider:        "gitlab",
				Repositories:    "example/foo",
				SshKey:          "-dummy-",
				GitLabServerURL: "https://my-gle.example.com",
			},
			want: map[string][]string{
				"git@my-gle.example.com:example/foo.git": {
					"git@my-gle.example.com:example/foo",
					"ssh://git@my-gle.example.com/example/foo.git",
					"ssh://git@my-gle.example.com/example/foo",
					"https://my-gle.example.com/example/foo.git",
					"https://my-gle.example.com/example/foo",
				},
			},
		},
		{
			name: "bitbucket cloud everything to https",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "*/*",
				SshKey:             "",
				BitbucketServerURL: "https://bitbucket.org",
			},
			want: map[string][]string{
				"https://bitbucket.org/": {
					"git@bitbucket.org:",
					"ssh://git@bitbucket.org/",
				},
			},
		},
		{
			name: "bitbucket cloud everything to ssh",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "*/*",
				SshKey:             "-dummy-",
				BitbucketServerURL: "https://bitbucket.org",
			},
			want: map[string][]string{
				"git@bitbucket.org:": {
					"ssh://git@bitbucket.org/",
					"https://bitbucket.org/",
				},
			},
		},
		{
			name: "bitbucket cloud org to https",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "example/*",
				SshKey:             "",
				BitbucketServerURL: "https://bitbucket.org",
			},
			want: map[string][]string{
				"https://bitbucket.org/example/": {
					"git@bitbucket.org:example/",
					"ssh://git@bitbucket.org/example/",
				},
			},
		},
		{
			name: "bitbucket cloud org to ssh",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "example/*",
				SshKey:             "-dummy-",
				BitbucketServerURL: "https://bitbucket.org",
			},
			want: map[string][]string{
				"git@bitbucket.org:example/": {
					"ssh://git@bitbucket.org/example/",
					"https://bitbucket.org/example/",
				},
			},
		},
		{
			name: "bitbucket cloud repo to https",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "example/foo",
				SshKey:             "",
				BitbucketServerURL: "https://bitbucket.org",
			},
			want: map[string][]string{
				"https://bitbucket.org/example/foo.git": {
					"git@bitbucket.org:example/foo.git",
					"ssh://git@bitbucket.org/example/foo.git",
				},
			},
		},
		{
			name: "bitbucket cloud some repos to ssh",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "example/foo",
				SshKey:             "-dummy-",
				BitbucketServerURL: "https://bitbucket.org",
			},
			want: map[string][]string{
				"git@bitbucket.org:example/foo.git": {
					"ssh://git@bitbucket.org/example/foo.git",
					"https://bitbucket.org/example/foo.git",
				},
			},
		},
		{
			name: "bitbucket enterprise everything to https",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "*/*",
				SshKey:             "",
				BitbucketServerURL: "https://my-bbs.example.com",
			},
			want: map[string][]string{
				"https://my-bbs.example.com/": {
					"git@my-bbs.example.com:",
					"ssh://git@my-bbs.example.com/",
				},
			},
		},
		{
			name: "bitbucket enterprise everything to ssh",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "*/*",
				SshKey:             "-dummy-",
				BitbucketServerURL: "https://my-bbs.example.com",
			},
			want: map[string][]string{
				"git@my-bbs.example.com:": {
					"ssh://git@my-bbs.example.com/",
					"https://my-bbs.example.com/",
				},
			},
		},
		{
			name: "bitbucket enterprise org to https",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "example/*",
				SshKey:             "",
				BitbucketServerURL: "https://my-bbs.example.com",
			},
			want: map[string][]string{
				"https://my-bbs.example.com/example/": {
					"git@my-bbs.example.com:example/",
					"ssh://git@my-bbs.example.com/example/",
				},
			},
		},
		{
			name: "bitbucket enterprise org to ssh",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "example/*",
				SshKey:             "-dummy-",
				BitbucketServerURL: "https://my-bbs.example.com",
			},
			want: map[string][]string{
				"git@my-bbs.example.com:example/": {
					"ssh://git@my-bbs.example.com/example/",
					"https://my-bbs.example.com/example/",
				},
			},
		},
		{
			name: "bitbucket enterprise repo to https",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "example/foo",
				SshKey:             "",
				BitbucketServerURL: "https://my-bbs.example.com",
			},
			want: map[string][]string{
				"https://my-bbs.example.com/example/foo.git": {
					"git@my-bbs.example.com:example/foo.git",
					"ssh://git@my-bbs.example.com/example/foo.git",
				},
			},
		},
		{
			name: "bitbucket enterprise some repos to ssh",
			fields: fields{
				Provider:           "bitbucket",
				Repositories:       "example/foo",
				SshKey:             "-dummy-",
				BitbucketServerURL: "https://my-bbs.example.com",
			},
			want: map[string][]string{
				"git@my-bbs.example.com:example/foo.git": {
					"ssh://git@my-bbs.example.com/example/foo.git",
					"https://my-bbs.example.com/example/foo.git",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Provider:           tt.fields.Provider,
				Repositories:       tt.fields.Repositories,
				SshKey:             tt.fields.SshKey,
				GitHubServerURL:    tt.fields.GitHubServerURL,
				BitbucketServerURL: tt.fields.BitbucketServerURL,
				GitLabServerURL:    tt.fields.GitLabServerURL,
			}
			got, err := c.insteadOfURLs()
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
