package cmd

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cmd = &cobra.Command{
		Use:          "configure-git-global-credentials",
		Short:        "Configures global credentials for accessing Git repositories",
		Long:         "Configures global credentials for accessing Git repositories",
		SilenceUsage: true,
		RunE:         doConfigure,
	}
)

func init() {
	viper.AutomaticEnv()

	viper.SetEnvPrefix("INPUT")

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	// inputString, etc
}

func inputString(name string, value string, usage string) {
	cmd.Flags().String(name, value, usage)
	_ = viper.BindPFlag(name, cmd.Flags().Lookup(name))
}

func Execute() error {
	return cmd.Execute()
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
	_ = cliContext()
	return nil
}
