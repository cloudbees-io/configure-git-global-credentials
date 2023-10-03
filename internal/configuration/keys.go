package configuration

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"

	"gopkg.in/alessio/shellescape.v1"
)

//go:embed ssh_known_hosts.tmpl
var sshKnownHostsTemplate string

func GenerateSSHKey(ctx context.Context, tempDir string, inputKey string) (string, error) {
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return "", err
	}
	keyPath := filepath.Join(tempDir, "private_key")
	if err := os.WriteFile(keyPath, []byte(inputKey), 0600); err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		// remove inherited permissions on windows
		icacls, err := exec.LookPath("icacls.exe")
		if err != nil && !errors.Is(err, exec.ErrDot) {
			return "", fmt.Errorf("cannot find icacls.exe: %w", err)
		} else if errors.Is(err, exec.ErrDot) {
			if icacls, err = filepath.Abs(icacls); err != nil {
				return "", fmt.Errorf("cannot find icacls.exe: %w", err)
			}
		}

		c := exec.CommandContext(ctx, icacls, keyPath, "/grant:r", os.Getenv("USERDOMAIN")+"\\"+os.Getenv("USERNAME")+":F")
		c.Dir = os.Getenv(tempDir)

		if err = c.Start(); err != nil {
			return "", err
		}

		if err = c.Wait(); err != nil {
			return "", err
		}

		c = exec.CommandContext(ctx, icacls, keyPath, "/inheritance:r")
		c.Dir = os.Getenv(tempDir)

		if err = c.Start(); err != nil {
			return "", err
		}

		if err = c.Wait(); err != nil {
			return "", err
		}
	}

	return keyPath, nil
}

func GenerateSSHCommand(sshKeyPath string, sshStrict bool, sshKnownHostsPath string) (string, error) {
	cmd := fmt.Sprintf("ssh -i %s", shellescape.Quote(sshKeyPath))
	if sshStrict {
		cmd = cmd + " -o StrictHostKeyChecking=yes -o CheckHostIP=no"
	}
	cmd = cmd + " -o UserKnownHostsFile=" + sshKnownHostsPath
	return cmd, nil
}

func GenerateSSHKnownHosts(home string, tempDir string, inputKnownHosts string) (_ string, retErr error) {
	tmpl := template.New("ssh_known_hosts")
	tmpl, err := tmpl.Parse(sshKnownHostsTemplate)
	if err != nil {
		return "", err
	}

	userKnownHostsPath := filepath.Join(home, ".ssh", "known_hosts")
	userKnownHosts := ""
	if info, err := os.Stat(userKnownHostsPath); err == nil && !info.IsDir() {
		if bytes, err := os.ReadFile(userKnownHostsPath); err == nil {
			userKnownHosts = string(bytes)
		}
		// if we couldn't read it, treat it as an empty file
	}

	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return "", err
	}
	knownHostsPath := filepath.Join(tempDir, "known_hosts")
	f, err := os.Create(knownHostsPath)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil && retErr == nil {
			retErr = err
		}
	}(f)

	err = tmpl.Execute(f, struct {
		UserKnownHosts     string
		UserKnownHostsPath string
		SSHKnownHosts      string
	}{
		UserKnownHosts:     userKnownHosts,
		UserKnownHostsPath: userKnownHostsPath,
		SSHKnownHosts:      inputKnownHosts,
	})
	return knownHostsPath, err
}
