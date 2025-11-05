package configuration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/config"
	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sshKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACB5tesp0633JJ+Q2hfpUXljwtBX263Tq9ENr76NdZ9e3wAAAKAFw5AuBcOQ
LgAAAAtzc2gtZWQyNTUxOQAAACB5tesp0633JJ+Q2hfpUXljwtBX263Tq9ENr76NdZ9e3w
AAAEApe1n3xwD4plUvs5E82QSBggtUz1M6HiiaVEYWp7ybpnm16ynTrfckn5DaF+lReWPC
0FfbrdOr0Q2vvo11n17fAAAAFnlvdXJfZW1haWxAZXhhbXBsZS5jb20BAgMEBQYH
-----END OPENSSH PRIVATE KEY-----
`

func TestConfig_SSH_Apply(t *testing.T) {
	tempDir := t.TempDir()
	tests := []struct {
		name              string
		config            Config
		wantErr           bool
		expectedGitConfig string
	}{
		{
			name: "Test with SSH without ssh url",
			config: Config{
				Repositories: "github.com/user/repo",
				SshKey:       sshKey,
			},
			wantErr: true,
		},
		{
			name: "Test with SSH with ssh url",
			config: Config{
				Repositories: "ssh://github.com/user/repo",
				SshKey:       sshKey,
			},
			wantErr:           false,
			expectedGitConfig: fmt.Sprintf("[core]\n\tsshCommand = ssh -i %s/.cloudbees-configure-git-global-credentials/513ad3faba989cce/private_key -o UserKnownHostsFile=%s/.cloudbees-configure-git-global-credentials/513ad3faba989cce/known_hosts\n[url \"ssh://github.com/user/repo\"]\n\tinsteadOf = ssh://github.com/user/repo\n", tempDir, tempDir),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helperBinary(t, tempDir)

			tempGitConfigPath := ""
			// mock loadConfig helper function
			loadConfig = func(scope config.Scope) (_ *format.Config, _ string, retErr error) {

				tempGitConfigPath = filepath.Join(tempDir, ".gitconfig")
				_, err := os.Create(tempGitConfigPath)
				require.NoError(t, err)

				d := format.NewDecoder(bytes.NewReader([]byte{}))
				result := format.Config{}
				err = d.Decode(&result)
				require.NoError(t, err)

				return &result, tempGitConfigPath, nil
			}

			context := context.Background()
			err := tt.config.Apply(context)
			assert.Equal(t, err != nil, tt.wantErr)
			// if there is err no further checks
			if !tt.wantErr {
				content, err := os.ReadFile(tempGitConfigPath)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedGitConfig, string(content))
			}

		})
	}
}

func TestConfig_Git_Credentials_Helper_Apply(t *testing.T) {
	tempDir := t.TempDir()
	tests := []struct {
		name              string
		config            Config
		wantErr           bool
		expectedGitConfig string
	}{
		{
			name: "Test with invalid repository url",
			config: Config{
				Repositories: "github.com/user/repo",
			},
			wantErr: true,
		},
		{
			name: "Test with repository",
			config: Config{
				Repositories: "https://github.com/user/repo",
			},
			wantErr: false,
		},
		{
			name: "Test with multiple repositories",
			config: Config{
				// testcase supports both comma and space as separator
				Repositories: "https://github.com/user/repo1, https://github.com/user/repo2 https://github.com/user/repo3",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helperBinary(t, tempDir)

			tempGitConfigPath := ""
			// mock loadConfig helper function
			loadConfig = func(scope config.Scope) (_ *format.Config, _ string, retErr error) {

				tempGitConfigPath = filepath.Join(tempDir, ".gitconfig")
				_, err := os.Create(tempGitConfigPath)
				require.NoError(t, err)

				d := format.NewDecoder(bytes.NewReader([]byte{}))
				result := format.Config{}
				err = d.Decode(&result)
				require.NoError(t, err)

				return &result, tempGitConfigPath, nil
			}

			gitCredentialsHelperInvoked := false
			pathActual := ""
			gitConfigPathActual := ""
			cloudbeesApiURLActual := ""
			cloudbeesApiTokenActual := ""
			filterGitUrlsActual := []string{}
			// mock credentials helper function
			invokeGitCredentialsHelper = func(ctx context.Context, path, gitConfigPath, cloudbeesApiURL, cloudbeesApiToken string, filterGitUrls []string) error {
				gitCredentialsHelperInvoked = true
				pathActual = path
				gitConfigPathActual = gitConfigPath
				cloudbeesApiURLActual = cloudbeesApiURL
				cloudbeesApiTokenActual = cloudbeesApiToken
				filterGitUrlsActual = filterGitUrls
				return nil
			}

			context := context.Background()
			err := tt.config.Apply(context)
			assert.Equal(t, err != nil, tt.wantErr)

			if !tt.wantErr {
				assert.True(t, gitCredentialsHelperInvoked, "git credentials helper should be invoked")
				assert.Equal(t, cbGitCredentialsHelper, pathActual)
				assert.Equal(t, tempGitConfigPath, gitConfigPathActual)
				assert.Equal(t, tt.config.CloudBeesApiURL, cloudbeesApiURLActual)
				assert.Equal(t, tt.config.CloudBeesApiToken, cloudbeesApiTokenActual)
				repoUrlArr := []string{}
				repoUrlArr = append(repoUrlArr, tt.config.repositories()...)
				assert.Equal(t, repoUrlArr, filterGitUrlsActual)
			}

		})
	}
}
func helperBinary(t *testing.T, tempDir string) {
	binPath := filepath.Join(tempDir, cbGitCredentialsHelper)
	_, err := os.Create(binPath)
	// err := os.WriteFile(binPath, []byte("dummy git credential helper"))
	require.NoError(t, err)

	err = os.Setenv("PATH", filepath.Dir(binPath))
	require.NoError(t, err)

	os.Setenv("HOME", tempDir)
}
