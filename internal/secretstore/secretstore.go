// Package secretstore keeps integration credentials outside crawl databases,
// exports, diagnostics, MCP responses, and profile JSON.
package secretstore

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type Store interface {
	Put(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
	Available(context.Context) error
}

var referencePattern = regexp.MustCompile(`^secret_[a-zA-Z0-9][a-zA-Z0-9._-]{0,99}$`)

func ValidateReference(value string) error {
	if !referencePattern.MatchString(value) {
		return errors.New("secret reference is invalid")
	}
	return nil
}

type Memory struct {
	mu     sync.RWMutex
	values map[string][]byte
}

func NewMemory() *Memory                          { return &Memory{values: map[string][]byte{}} }
func (m *Memory) Available(context.Context) error { return nil }
func (m *Memory) Put(_ context.Context, ref string, value []byte) error {
	if err := ValidateReference(ref); err != nil {
		return err
	}
	if len(value) < 1 || len(value) > 64<<10 {
		return errors.New("secret value is outside supported bounds")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[ref] = append([]byte(nil), value...)
	return nil
}
func (m *Memory) Get(_ context.Context, ref string) ([]byte, error) {
	if err := ValidateReference(ref); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.values[ref]
	if !ok {
		return nil, errors.New("secret reference was not found")
	}
	return append([]byte(nil), value...), nil
}
func (m *Memory) Delete(_ context.Context, ref string) error {
	if err := ValidateReference(ref); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.values, ref)
	return nil
}

type unavailable struct{ reason string }

func (u unavailable) Available(context.Context) error             { return errors.New(u.reason) }
func (u unavailable) Put(context.Context, string, []byte) error   { return errors.New(u.reason) }
func (u unavailable) Get(context.Context, string) ([]byte, error) { return nil, errors.New(u.reason) }
func (u unavailable) Delete(context.Context, string) error        { return errors.New(u.reason) }

type commandStore struct{ platform, binary string }

func NewPlatform() Store {
	switch runtime.GOOS {
	case "linux":
		if path, err := exec.LookPath("secret-tool"); err == nil {
			return commandStore{"linux", path}
		}
		return unavailable{"Linux Secret Service is unavailable (secret-tool not found)"}
	case "darwin":
		if path, err := exec.LookPath("security"); err == nil {
			return commandStore{"darwin", path}
		}
		return unavailable{"macOS Keychain command is unavailable"}
	case "windows":
		if path, err := exec.LookPath("powershell.exe"); err == nil {
			return commandStore{"windows", path}
		}
		return unavailable{"Windows Credential Manager bridge is unavailable"}
	default:
		return unavailable{"the platform has no supported secure credential store"}
	}
}
func (s commandStore) Available(context.Context) error {
	if s.binary == "" {
		return errors.New("secure credential store is unavailable")
	}
	return nil
}
func (s commandStore) Put(ctx context.Context, ref string, value []byte) error {
	if err := ValidateReference(ref); err != nil {
		return err
	}
	if len(value) < 1 || len(value) > 64<<10 {
		return errors.New("secret value is outside supported bounds")
	}
	var command *exec.Cmd
	switch s.platform {
	case "linux":
		command = exec.CommandContext(ctx, s.binary, "store", "--label=SEO Screaming Toad: "+ref, "service", "seo-screaming-toad", "reference", ref)
		command.Stdin = strings.NewReader(string(value))
	case "darwin":
		command = exec.CommandContext(ctx, s.binary, "add-generic-password", "-U", "-s", "seo-screaming-toad", "-a", ref, "-w")
		command.Stdin = strings.NewReader(string(value))
	case "windows":
		command = exec.CommandContext(ctx, s.binary, "-NoProfile", "-NonInteractive", "-Command", windowsPut, ref)
		command.Stdin = strings.NewReader(string(value))
	}
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("secure credential store write failed: %s", boundedError(output))
	}
	return nil
}
func (s commandStore) Get(ctx context.Context, ref string) ([]byte, error) {
	if err := ValidateReference(ref); err != nil {
		return nil, err
	}
	var command *exec.Cmd
	switch s.platform {
	case "linux":
		command = exec.CommandContext(ctx, s.binary, "lookup", "service", "seo-screaming-toad", "reference", ref)
	case "darwin":
		command = exec.CommandContext(ctx, s.binary, "find-generic-password", "-s", "seo-screaming-toad", "-a", ref, "-w")
	case "windows":
		command = exec.CommandContext(ctx, s.binary, "-NoProfile", "-NonInteractive", "-Command", windowsGet, ref)
	}
	output, err := command.Output()
	if err != nil {
		return nil, errors.New("secure credential lookup failed")
	}
	output = []byte(strings.TrimRight(string(output), "\r\n"))
	if len(output) == 0 || len(output) > 64<<10 {
		return nil, errors.New("secure credential is missing or outside supported bounds")
	}
	return output, nil
}
func (s commandStore) Delete(ctx context.Context, ref string) error {
	if err := ValidateReference(ref); err != nil {
		return err
	}
	var command *exec.Cmd
	switch s.platform {
	case "linux":
		command = exec.CommandContext(ctx, s.binary, "clear", "service", "seo-screaming-toad", "reference", ref)
	case "darwin":
		command = exec.CommandContext(ctx, s.binary, "delete-generic-password", "-s", "seo-screaming-toad", "-a", ref)
	case "windows":
		command = exec.CommandContext(ctx, s.binary, "-NoProfile", "-NonInteractive", "-Command", windowsDelete, ref)
	}
	if err := command.Run(); err != nil {
		return errors.New("secure credential deletion failed")
	}
	return nil
}
func boundedError(value []byte) string {
	message := strings.TrimSpace(string(value))
	if len(message) > 500 {
		message = message[:500]
	}
	if message == "" {
		return "no diagnostic returned"
	}
	return message
}

const windowsPut = `$r=$args[0];$s=[Console]::In.ReadToEnd();Add-Type -AssemblyName System.Runtime.WindowsRuntime;$v=New-Object Windows.Security.Credentials.PasswordVault;$v.Add((New-Object Windows.Security.Credentials.PasswordCredential('seo-screaming-toad',$r,$s)))`
const windowsGet = `$r=$args[0];Add-Type -AssemblyName System.Runtime.WindowsRuntime;$v=New-Object Windows.Security.Credentials.PasswordVault;$c=$v.Retrieve('seo-screaming-toad',$r);$c.RetrievePassword();[Console]::Out.Write($c.Password)`
const windowsDelete = `$r=$args[0];Add-Type -AssemblyName System.Runtime.WindowsRuntime;$v=New-Object Windows.Security.Credentials.PasswordVault;$c=$v.Retrieve('seo-screaming-toad',$r);$v.Remove($c)`
