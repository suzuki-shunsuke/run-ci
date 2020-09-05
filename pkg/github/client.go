package github

import (
	"context"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

type Client struct {
	Client *github.Client
}

type ParamsNew struct {
	Token string
}

func New(ctx context.Context, params ParamsNew) Client {
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: params.Token},
	))
	return Client{
		Client: github.NewClient(tc),
	}
}

type ParamsListPRs struct {
	Owner string
	Repo  string
	Base  string
	Head  string
}

func (client Client) ListPRs(ctx context.Context, params ParamsListPRs) ([]*github.PullRequest, *github.Response, error) {
	return client.Client.PullRequests.List(ctx, params.Owner, params.Repo, &github.PullRequestListOptions{
		Base: params.Base,
		Head: params.Head,
	})
}

type ParamsGetRef struct {
	Owner string
	Repo  string
	Ref   string
}

func (client Client) GetRef(ctx context.Context, params ParamsGetRef) (*github.Reference, *github.Response, error) {
	return client.Client.Git.GetRef(ctx, params.Owner, params.Repo, params.Ref)
}

type ParamsGetCommit struct {
	Owner string
	Repo  string
	SHA   string
}

func (client Client) GetCommit(ctx context.Context, params ParamsGetCommit) (*github.Commit, *github.Response, error) {
	return client.Client.Git.GetCommit(ctx, params.Owner, params.Repo, params.SHA)
}

type ParamsCreateEmptyCommit struct {
	Owner     string
	Repo      string
	CommitMsg string
	Parent    *github.Commit
}

func (client Client) CreateEmptyCommit(ctx context.Context, params ParamsCreateEmptyCommit) (*github.Commit, *github.Response, error) {
	msg := params.CommitMsg
	return client.Client.Git.CreateCommit(ctx, params.Owner, params.Repo, &github.Commit{
		Tree:    params.Parent.Tree,
		Message: &msg,
		Parents: []*github.Commit{params.Parent},
	})
}

type ParamsUpdateRef struct {
	Owner string
	Repo  string
	Ref   string
	SHA   string
}

func (client Client) UpdateRef(ctx context.Context, params ParamsUpdateRef) (*github.Reference, *github.Response, error) {
	ref := params.Ref
	sha := params.SHA
	return client.Client.Git.UpdateRef(ctx, params.Owner, params.Repo, &github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			SHA: &sha,
		},
	}, true)
}
