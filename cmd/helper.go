package cmd

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/cloudbees-io/configure-git-global-credentials/internal/helper"
	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/spf13/cobra"
)

var (
	helperCmd = &cobra.Command{
		Use:                "credential-helper",
		Short:              "Implements the Git Credentials Helper API",
		Long:               "Implements the Git Credentials Helper API",
		SilenceUsage:       true,
		DisableSuggestions: true,
		Run: func(cmd *cobra.Command, args []string) {
			// From the https://git-scm.com/docs/gitcredentials specification
			//
			// If a helper receives any other operation, it should silently ignore the request.
			// This leaves room for future operations to be added (older helpers will just
			// ignore the new requests).
		},
	}
	eraseCmd = &cobra.Command{
		Use:          "erase",
		Short:        "Remove a matching credential, if any, from the helper’s storage",
		Long:         "Remove a matching credential, if any, from the helper’s storage",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			// we do not support the erase operation
		},
	}
	storeCmd = &cobra.Command{
		Use:          "store",
		Short:        "Store the credential, if applicable to the helper",
		Long:         "Store the credential, if applicable to the helper",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			// we do not support the store operation
		},
	}
	getCmd = &cobra.Command{
		Use:          "get",
		Short:        "Return matching credential, if any exists",
		Long:         "Return matching credential, if any exists",
		SilenceUsage: true,
		RunE:         doGet,
	}

	helperConfigFile string
)

func init() {
	helperCmd.AddCommand(getCmd, eraseCmd, storeCmd)
	helperCmd.PersistentFlags().StringVarP(&helperConfigFile, "config-file", "c", "", "path to the helper configuration file to use")
}

func doGet(command *cobra.Command, args []string) error {
	_ = cliContext()

	if helperConfigFile == "" {
		self, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot infer config file from executable name: %w", err)
		}
		helperConfigFile = self + ".cfg"
	}

	bs, err := os.ReadFile(helperConfigFile)
	if err != nil {
		return fmt.Errorf("could not read configuration from %s: %w", helperConfigFile, err)
	}

	cfg := format.Config{}

	d := format.NewDecoder(bytes.NewReader(bs))

	if err := d.Decode(&cfg); err != nil {
		return fmt.Errorf("could not parse configuration file %s: %w", helperConfigFile, err)
	}

	r := bufio.NewReader(os.Stdin)

	req, err := helper.ReadCredential(r)
	if err != nil {
		return fmt.Errorf("could not read credentials request: %w", err)
	}

	section := cfg.Section(req.Protocol)

	target := (&transport.Endpoint{
		Host: req.Host,
		Path: req.Path,
	}).String()

	var closest *format.Subsection
	for _, ss := range section.Subsections {
		if strings.HasPrefix(target, ss.Name) {
			if closest == nil || (len(closest.Name) < len(ss.Name)) {
				closest = ss
			}
		}
	}

	if closest == nil {
		// nothing to contribute
		return nil
	}

	rsp := &helper.GitCredential{}

	if closest.HasOption("username") {
		rsp.Username = closest.Option("username")
	}

	if closest.HasOption("password") {
		if b, err := base64.StdEncoding.DecodeString(closest.Option("password")); err == nil {
			rsp.Password = string(b)
		} else {
			return err
		}
	}

	// TODO SDP-6415 handle fetching credentials

	w := bufio.NewWriter(os.Stdout)

	if _, err = rsp.WriteTo(w); err != nil {
		return err
	}

	return w.Flush()
}
