package configuration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
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
	// SshKey SSH key used to fetch the repositories
	SshKey string `mapstructure:"ssh-key"`
	// SshKnownHosts Known hosts in addition to the user and global host key database
	SshKnownHosts string `mapstructure:"ssh-known-hosts"`
	// SshStrict Whether to perform strict host key checking
	SshStrict bool `mapstructure:"ssh-strict"`
}

const (
	tokenEnv               = "CLOUDBEES_API_TOKEN"
	cbGitCredentialsHelper = "git-credential-cloudbees"
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

func (c *Config) validate() error {
	if c.ssh() {
		for _, sshUrl := range c.repositories() {
			if !isSSHURL(sshUrl) {
				return fmt.Errorf("invalid SSH URL: %s", sshUrl)
			}
		}
	} else {
		for _, repoUrl := range c.repositories() {
			parsedURL, err := url.Parse(repoUrl)
			if err != nil {
				return fmt.Errorf("invalid repository %q: %w", repoUrl, err)
			}
			if !parsedURL.IsAbs() || parsedURL.Host == "" {
				return fmt.Errorf("invalid repository URL %q provided, expects full clone URL", repoUrl)
			}
		}
	}

	return nil
}

const userAndHostRegex = `([a-zA-Z][-a-zA-Z0-9_]*@)?[a-z0-9][-a-z0-9_\.]*`

// Matches SSH URLs in the format: ssh://user@host[:port]/path
var sshURLRegexScheme = regexp.MustCompile(fmt.Sprintf(`^ssh://%s(:|/)(/?[\w_\-\.~]+)*$`, userAndHostRegex))

// Matches SSH URLs in the format: user@host:path
var sshURLRegexScp = regexp.MustCompile(fmt.Sprintf(`^%s:/?[\w_\-\.~]+(/?[\w_\-\.~]+)*$`, userAndHostRegex))

func isSSHURL(urlStr string) bool {
	return sshURLRegexScheme.MatchString(urlStr) || sshURLRegexScp.MatchString(urlStr)
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
	for _, sshUrl := range c.repositories() {
		urlSection := cfg.Section("url")
		urlHostEntry := urlSection.Subsection(sshUrl)
		urlHostEntry.SetOption("insteadOf", sshUrl)
	}

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

	if err := c.validate(); err != nil {
		return err
	}

	fmt.Println("🔄 Parsing existing Git global config ...")

	_, cfgPath, err := loadConfig(config.GlobalScope)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Git global config at %s parsed\n", cfgPath)

	repoUrlArr := c.repositories()
	filterUrl := make([]string, 0, len(repoUrlArr))
	filterUrl = append(filterUrl, repoUrlArr...)

	if c.ssh() {
		return c.setupSsh(ctx)
	} else {
		return invokeGitCredentialsHelper(ctx, cbGitCredentialsHelper, cfgPath, c.CloudBeesApiURL, c.CloudBeesApiToken, filterUrl)
	}
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
	return re.Split(strings.TrimSpace(c.Repositories), -1)
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
