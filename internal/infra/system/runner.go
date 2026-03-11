package system

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Runner abstracts shell execution and platform checks for infra adapters.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
	LookPath(file string) (string, error)
	GOOS() string
}

type ExecRunner struct{}

func NewExecRunner() Runner {
	return ExecRunner{}
}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (ExecRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (ExecRunner) GOOS() string {
	return runtime.GOOS
}
