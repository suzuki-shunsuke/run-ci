package git

import (
	"context"

	"github.com/suzuki-shunsuke/run-ci/pkg/execute"
)

type Executor interface {
	Run(context.Context, execute.Params) error
}

type Client struct {
	Executor  Executor
	UserName  string
	UserEmail string
}

type ParamsNew struct {
	UserName  string
	UserEmail string
	Executor  Executor
}

func New(params ParamsNew) Client {
	return Client{
		Executor:  params.Executor,
		UserName:  params.UserName,
		UserEmail: params.UserEmail,
	}
}

func (client Client) env() []string {
	return []string{
		"GIT_AUTHOR_NAME=" + client.UserName,
		"GIT_AUTHOR_EMAIL=" + client.UserEmail,
		"GIT_COMMITTER_NAME=" + client.UserName,
		"GIT_COMMITTER_EMAIL=" + client.UserEmail,
	}
}

func (client Client) Fetch(ctx context.Context, remote, branch string) error {
	return client.Executor.Run(ctx, execute.Params{
		Cmd: "git",
		Args: []string{
			"fetch", remote, branch,
		},
	})
}

func (client Client) Checkout(ctx context.Context, branch string) error {
	return client.Executor.Run(ctx, execute.Params{
		Cmd: "git",
		Args: []string{
			"checkout", branch,
		},
	})
}

func (client Client) CommitEmpty(ctx context.Context, msg string) error {
	return client.Executor.Run(ctx, execute.Params{
		Cmd: "git",
		Args: []string{
			"commit", "--allow-empty", "-m", msg,
		},
		Envs: client.env(),
	})
}

func (client Client) Push(ctx context.Context, branch string) error {
	return client.Executor.Run(ctx, execute.Params{
		Cmd: "git",
		Args: []string{
			"push", "origin", branch,
		},
	})
}

func (client Client) Reset(ctx context.Context) error {
	return client.Executor.Run(ctx, execute.Params{
		Cmd: "git",
		Args: []string{
			"reset", "HEAD~1", "--hard",
		},
	})
}

func (client Client) PushForce(ctx context.Context, branch string) error {
	return client.Executor.Run(ctx, execute.Params{
		Cmd: "git",
		Args: []string{
			"push", "origin", branch, "--force",
		},
	})
}

func (client Client) Merge(ctx context.Context, branch string) error {
	return client.Executor.Run(ctx, execute.Params{
		Cmd: "git",
		Args: []string{
			"merge", "--no-edit", branch,
		},
		Envs: client.env(),
	})
}
