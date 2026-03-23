package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/cloudbees-io/configure-git-global-credentials/internal/configuration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "configure-git-global-credentials",
		Short: "Configures global credentials for accessing Git repositories",
		Long:  "Configures global credentials for accessing Git repositories",
	}
	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "Configures the global git credentials",
		Long:  "Configures the global git credentials",
		RunE:  doConfigure,
	}
	// Ensure warning is printed only once
	warningOnce sync.Once
)

func init() {
	viper.AutomaticEnv()

	viper.SetEnvPrefix("INPUT")

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	inputString("provider", "", "SCM provider that is hosting the repositories")
	inputString("repositories", "", "Whitespace and/or comma separated list of repository names with owner")
	inputString("cloudbees-api-token", "", "CloudBees API token used to fetch authentication")
	inputString("cloudbees-api-url", "", "CloudBees API root URL to fetch authentication from")
	inputString("token", "", "Personal access token (PAT) used to fetch the repositories")
	inputString("ssh-key", "", "SSH key used to fetch the repositories")
	inputString("ssh-known-hosts", "", "Known hosts in addition to the user and global host key database")
	inputBool("ssh-strict", true, "Whether to perform strict host key checking")
	inputString("github-server-url", "", "The base URL for the GitHub instance that you are trying to clone from")
	inputString("bitbucket-server-url", "", "The base URL for the Bitbucket instance that you are trying to clone from")
	inputString("gitlab-server-url", "", "The base URL for the GitLab instance that you are trying to clone from")

	rootCmd.AddCommand(helperCmd, configureCmd)
}

func inputString(name string, value string, usage string) {
	configureCmd.Flags().String(name, value, usage)
	_ = viper.BindPFlag(name, configureCmd.Flags().Lookup(name))
}

func inputBool(name string, value bool, usage string) {
	configureCmd.Flags().Bool(name, value, usage)
	_ = viper.BindPFlag(name, configureCmd.Flags().Lookup(name))
}

func Execute() error {
	return rootCmd.Execute()
}

func cliContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel() // exit gracefully
		<-c
		os.Exit(1) // exit immediately on 2nd signal
	}()
	return ctx
}

func doConfigure(command *cobra.Command, args []string) error {
	// Display v1 deprecation warning (only once)
	warningOnce.Do(printV1DeprecationWarning)

	ctx := cliContext()

	var cfg configuration.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	return cfg.Apply(ctx)
}

func printV1DeprecationWarning() {
	warningMessage := `
==================================================================================================
                                        ⚠️ DEPRECATION WARNING ⚠️
==================================================================================================

You are using the DEPRECATED v1 version of the configure-git-global-credentials action.

Version v1 is no longer maintained and will be removed in the future.

⚡ Please migrate to v2 as soon as possible.

📖 Migration Guide:
https://docs.cloudbees.com/docs/cloudbees-platform/latest/source-code-management/migrate-v1-to-v2

Update your workflow file:
  Change: uses: cloudbees-io/configure-git-global-credentials@v1
  To:     uses: cloudbees-io/configure-git-global-credentials@v2

==================================================================================================
`
	fmt.Fprint(os.Stderr, warningMessage)
}
