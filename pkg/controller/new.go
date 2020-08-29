package controller

import (
	"context"

	"github.com/google/go-github/v32/github"
	"github.com/suzuki-shunsuke/run-ci/pkg/config"
	gh "github.com/suzuki-shunsuke/run-ci/pkg/github"
)

type Controller struct {
	Git    Git
	GitHub GitHub
	Expr   Expr
	Config config.Config
}

type Git interface {
	Fetch(ctx context.Context, branch string) error
	Checkout(ctx context.Context, branch string) error
	CommitEmpty(ctx context.Context, msg string) error
	Push(ctx context.Context, branch string) error
	PushForce(ctx context.Context, branch string) error
	Reset(ctx context.Context) error
	Merge(ctx context.Context, branch string) error
}

type GitHub interface {
	ListPRs(ctx context.Context, params gh.ParamsListPRs) ([]*github.PullRequest, *github.Response, error)
	GetRef(ctx context.Context, params gh.ParamsGetRef) (*github.Reference, *github.Response, error)
	GetCommit(ctx context.Context, params gh.ParamsGetCommit) (*github.Commit, *github.Response, error)
	CreateEmptyCommit(ctx context.Context, params gh.ParamsCreateEmptyCommit) (*github.Commit, *github.Response, error)
	UpdateRef(ctx context.Context, params gh.ParamsUpdateRef) (*github.Reference, *github.Response, error)
}

type Expr interface {
	Match(params interface{}) (bool, error)
}
