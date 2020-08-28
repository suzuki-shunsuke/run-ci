package config

import "os"

type Config struct {
	Owner          string
	Repo           string
	EmptyCommitMsg string `yaml:"empty_commit_msg"`
	Expr           string
	GitHubToken    string        `yaml:"github_token"`
	GitCommand     GitCommand    `yaml:"git_command"`
	PullRequests   []PullRequest `yaml:"pull_requests"`
}

type PullRequest struct {
	Head string
	Base string
	Expr string
}

type GitCommand struct {
	UserName  string `yaml:"user_name"`
	UserEmail string `yaml:"user_email"`
	Use       bool
}

func SetDefault(cfg Config) Config {
	if cfg.EmptyCommitMsg == "" {
		cfg.EmptyCommitMsg = "[ci skip]"
	}
	if cfg.GitCommand.UserName == "" {
		cfg.GitCommand.UserName = "run-ci"
	}
	if cfg.GitCommand.UserEmail == "" {
		cfg.GitCommand.UserEmail = "run-ci@example.com"
	}
	if cfg.Expr == "" {
		cfg.Expr = "true"
	}
	return cfg
}

func SetEnv(cfg Config) Config {
	if cfg.GitHubToken == "" {
		cfg.GitHubToken = os.Getenv("GITHUB_TOKEN")
	}
	if cfg.GitHubToken == "" {
		cfg.GitHubToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	}
	return cfg
}
