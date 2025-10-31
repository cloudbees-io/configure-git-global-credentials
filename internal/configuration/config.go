package configuration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/cloudbees-io/configure-git-global-credentials/internal"
	"github.com/go-git/go-git/v5/config"
	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"golang.org/x/crypto/ssh"
)

// Config holds the configuration of the authentication to be applied
type Config struct {
	// Repositories whitespace and/or comma separated list of repository names with owner
	Repositories string
	// CloudBees API token used to fetch authentication
	CloudBeesApiToken string `mapstructure:"cloudbees-api-token"`
	// CloudBees API root URL to fetch authentication from
	CloudBeesApiURL string `mapstructure:"cloudbees-api-url"`
	// Personal access token (PAT) used to fetch the repositories
	Token string
	// SshKey SSH key used to fetch the repositories
	SshKey string `mapstructure:"ssh-key"`
	// SshKnownHosts Known hosts in addition to the user and global host key database
	SshKnownHosts string `mapstructure:"ssh-known-hosts"`
	// SshStrict Whether to perform strict host key checking
	SshStrict bool `mapstructure:"ssh-strict"`
	// GitHubServerURL the base URL for the GitHub instance that you are trying to clone from
	GitHubServerURL string `mapstructure:"github-server-url"`
	// BitbucketServerURL the base URL for the Bitbucket instance that you are trying to clone from
	BitbucketServerURL string `mapstructure:"bitbucket-server-url"`
	// GitLabServerURL the base URL for the GitLab instance that you are trying to clone from
	// GitLabServerURL string `mapstructure:"gitlab-server-url"`
}

const (
	tokenEnv                   = "CLOUDBEES_API_TOKEN"
	cbGitCredentialsHelperPath = "git-credential-cloudbees"
)

var loadConfig = func(scope config.Scope) (_ *format.Config, _ string, retErr error) {
	paths, err := config.Paths(scope)
	if err != nil {
		return nil, "", err
	}

	result := format.Config{}

	for _, file := range paths {
		f, err := os.ReadFile(file)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, "", err
		}

		d := format.NewDecoder(bytes.NewReader(f))

		if err := d.Decode(&result); err != nil {
			return nil, "", err
		}

		return &result, file, nil
	}

	return &result, paths[0], nil
}

func (c *Config) populateDefaults(ctx context.Context) error {
	c.Token = strings.TrimSpace(c.Token)
	c.SshKey = strings.TrimSpace(c.SshKey)

	if c.Token != "" && c.SshKey != "" {
		return fmt.Errorf("input parameters 'token' and 'ssh-key' are mutually exclusive")
	}

	return nil
}

func (c *Config) setupSsh(ctx context.Context) error {

	cfg, cfgPath, err := loadConfig(config.GlobalScope)
	if err != nil {
		return err
	}
	homePath := os.Getenv("HOME")
	actionPath := filepath.Join(homePath, ".cloudbees-configure-git-global-credentials", c.uniqueId())
	if err := os.MkdirAll(actionPath, os.ModePerm); err != nil {
		return err
	}

	// check if the SSH key looks to be a base64 encoded private key that the user forgot to decode
	if decoded, err := base64.StdEncoding.DecodeString(c.SshKey); err == nil {
		sshKey := string(decoded)
		if err == nil && strings.Contains(sshKey, "-----BEGIN") && strings.Contains(sshKey, "PRIVATE KEY-----") {
			fmt.Println("✅ Base64 decoded SSH key")
			c.SshKey = sshKey
		}
	}

	if _, err := ssh.ParseRawPrivateKey([]byte(c.SshKey)); err != nil {
		fmt.Println("❌ Could not parse supplied SSH key")
		return fmt.Errorf("could not parse supplied SSH key: %w", err)
	}
	fmt.Println("🔄 Installing SSH private key ...")

	sshKeyPath, err := GenerateSSHKey(ctx, actionPath, c.SshKey)
	if err != nil {
		return err
	}

	sshKnownHostsPath, err := GenerateSSHKnownHosts(homePath, actionPath, c.SshKnownHosts)
	if err != nil {
		return err
	}

	sshCommand, err := GenerateSSHCommand(sshKeyPath, c.SshStrict, sshKnownHostsPath)
	if err != nil {
		return err
	}

	cfg.Section("core").SetOption("sshCommand", sshCommand)

	fmt.Println("✅ SSH private key installed")
	fmt.Printf("🔄 Updating %s ...\n", cfgPath)

	var b bytes.Buffer
	if err := format.NewEncoder(&b).Encode(cfg); err != nil {
		return err
	}

	if err := os.WriteFile(cfgPath, b.Bytes(), 0666); err != nil {
		return err
	}

	fmt.Printf("✅ Git global config at %s updated\n", cfgPath)

	return nil
}

// Apply applies the configuration to the Git Global config
func (c *Config) Apply(ctx context.Context) error {

	if err := c.populateDefaults(ctx); err != nil {
		return err
	}

	fmt.Println("🔄 Parsing existing Git global config ...")

	_, cfgPath, err := loadConfig(config.GlobalScope)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Git global config at %s parsed\n", cfgPath)

	repoUrlArr := c.repositories()
	fmt.Printf("repoUrlArr %v\n", repoUrlArr)
	filterUrl := make([]string, 0, len(repoUrlArr))
	filterUrl = append(filterUrl, repoUrlArr...)
	err = invokeGitCredentialsHelper(ctx, cbGitCredentialsHelperPath, cfgPath, c.CloudBeesApiURL, c.CloudBeesApiToken, filterUrl)
	if err != nil {
		return err
	}

	if c.ssh() {
		return c.setupSsh(ctx)
	}

	return nil
}

var invokeGitCredentialsHelper = func(ctx context.Context, path, gitConfigPath, cloudbeesApiURL, cloudbeesApiToken string, filterGitUrls []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	helperConfig := filepath.Join(homeDir, ".git-credential-cloudbees-config")

	filterUrlArgs := []string{}

	filterUrlArgs = append(filterUrlArgs, "init")
	filterUrlArgs = append(filterUrlArgs, "--config", helperConfig)
	filterUrlArgs = append(filterUrlArgs, "--cloudbees-api-token-env-var", tokenEnv)
	filterUrlArgs = append(filterUrlArgs, "--cloudbees-api-url", cloudbeesApiURL)
	filterUrlArgs = append(filterUrlArgs, "--git-config-file-path", gitConfigPath)
	for _, filterGitUrl := range filterGitUrls {
		filterUrlArgs = append(filterUrlArgs, "--filter-git-urls", filterGitUrl)
	}
	cmd := exec.CommandContext(ctx, path, filterUrlArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	internal.Debug("%s", cmd.String())

	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", tokenEnv, cloudbeesApiToken))

	return cmd.Run()
}

func (c *Config) ssh() bool {
	return strings.TrimSpace(c.SshKey) != ""
}

func (c *Config) repositories() []string {
	if c.Repositories == "" || strings.TrimSpace(c.Repositories) == "" {
		if c.SshKey != "" {
			return []string{"*/*"}
		}
		return nil
	}
	re := regexp.MustCompile(`[ \t\r\n\f,]+`)
	return re.Split(c.Repositories, -1)
}

func (c *Config) uniqueId() string {
	r := c.repositories()
	sort.Stable(sort.StringSlice(r))
	h := sha256.New()
	for _, v := range r {
		h.Write([]byte(v))
		h.Write([]byte{0})
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)[0:16]
}

// func (c *Config) insteadOfURLs() (map[string][]string, error) {
// 	ssh := c.ssh()
// 	repos := c.repositories()
// 	result := make(map[string][]string, len(repos))
// 	for _, r := range repos {
// 		preferred, err := c.allURLs(ssh, r)
// 		if err != nil {
// 			return nil, err
// 		}
// 		alternate, err := c.allURLs(!ssh, r)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if len(preferred) > 1 {
// 			result[preferred[0]] = append(preferred[1:], alternate...)
// 		} else if len(preferred) == 1 {
// 			result[preferred[0]] = alternate
// 		}
// 	}
// 	return result, nil
// }

func findEventContext() map[string]interface{} {
	if eventPath, found := os.LookupEnv("CLOUDBEES_EVENT_PATH"); found {
		return safeLoadEventContext(eventPath)
	} else if homePath, found := os.LookupEnv("CLOUDBEES_HOME"); found {
		// TODO remove when CLOUDBEES_EVENT_PATH is exposed in the environment
		return safeLoadEventContext(filepath.Join(homePath, "event.json"))
	}
	return make(map[string]interface{})
}

// safeLoadEventContext attempts to load the event context from the JSON file at the supplied path always returning
// a (possibly empty) map.
func safeLoadEventContext(path string) map[string]interface{} {
	c, err := loadEventContext(path)
	if err != nil {
		return make(map[string]interface{})
	}
	return c
}

// loadEventContext attempts to load the event context from the JSON file at the supplied path.
func loadEventContext(path string) (map[string]interface{}, error) {
	var bytes []byte
	var err error

	if bytes, err = os.ReadFile(path); err != nil {
		// best effort
		return nil, err
	}

	var event map[string]interface{}
	if err = json.Unmarshal(bytes, &event); err != nil {
		// best effort
		return nil, err
	}

	return event, nil
}

func getStringFromMap(m map[string]interface{}, key string) (string, bool) {
	i, found := m[key]
	if !found {
		return "", false
	}
	if s, ok := i.(string); ok {
		return s, true
	}
	return "", false
}

func getBoolFromMap(m map[string]interface{}, key string) (bool, bool) {
	i, found := m[key]
	if !found {
		return false, false
	}
	if v, ok := i.(bool); ok {
		return v, true
	}
	return false, false
}

func getMapFromMap(m map[string]interface{}, key string) (map[string]interface{}, bool) {
	i, found := m[key]
	if !found {
		return map[string]interface{}{}, false
	}
	if v, ok := i.(map[string]interface{}); ok {
		return v, true
	}
	return map[string]interface{}{}, false
}

func copyFileHelper(dst string, src string) (err error) {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err2 := f.Close()
		if err2 != nil && err == nil {
			err = err2
		}
	}(s)

	if stat, err := os.Stat(dst); err == nil {
		// set up to force delete
		if err := os.Chmod(dst, stat.Mode()|0222); err != nil {
			return err
		}
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	// Create the destination file with default permission
	d, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0555)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err2 := f.Close()
		if err2 != nil && err == nil {
			err = err2
		}
	}(d)

	_, err = io.Copy(d, s)
	return err
}
