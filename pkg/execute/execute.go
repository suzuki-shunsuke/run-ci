package execute

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

type Executor struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Environ []string
}

func New() Executor {
	return Executor{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Environ: os.Environ(),
	}
}

type Params struct {
	Cmd        string
	Args       []string
	WorkingDir string
	Envs       []string
	Quiet      bool
	DryRun     bool
	Timeout    Timeout
}

type Timeout struct {
	Duration  time.Duration
	KillAfter time.Duration
}

func (exc Executor) Run(ctx context.Context, params Params) error {
	cmd := exec.Command(params.Cmd, params.Args...) //nolint:gosec
	cmd.Stdout = exc.Stdout
	cmd.Stderr = exc.Stderr
	cmd.Stdin = exc.Stdin
	cmd.Dir = params.WorkingDir

	cmd.Env = append(exc.Environ, params.Envs...) //nolint:gocritic
	if !params.Quiet {
		fmt.Fprintln(exc.Stderr, "+ "+params.Cmd+strings.Join(params.Args, " "))
	}
	if params.DryRun {
		return nil
	}
	runner := timeout.NewRunner(params.Timeout.KillAfter)
	runner.SetSigKillCaballback(func(targetID int) {
		fmt.Fprintf(exc.Stderr, "send SIGKILL to %d\n", targetID)
	})

	if params.Timeout.Duration > 0 {
		c, cancel := context.WithTimeout(ctx, params.Timeout.Duration)
		defer cancel()
		ctx = c
	}
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintf(exc.Stderr, "command is terminated by timeout: %d seconds\n", params.Timeout.Duration)
		}
	}()
	if err := runner.Run(ctx, cmd); err != nil {
		return ecerror.Wrap(err, cmd.ProcessState.ExitCode())
	}
	return nil
}
