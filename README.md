# run-ci

[![Build Status](https://github.com/suzuki-shunsuke/run-ci/workflows/CI/badge.svg)](https://github.com/suzuki-shunsuke/run-ci/actions)
[![Test Coverage](https://api.codeclimate.com/v1/badges/005f67bb7c2e12f59824/test_coverage)](https://codeclimate.com/github/suzuki-shunsuke/run-ci/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/run-ci)](https://goreportcard.com/report/github.com/suzuki-shunsuke/run-ci)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/run-ci.svg)](https://github.com/suzuki-shunsuke/run-ci)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/run-ci/master/LICENSE)

CLI tool to run CI automatically when pull request's base branch is updated.

## Overview

`run-ci` is a CLI tool to run CI when pull request's base branch is updated.

When the base branch is updated, the existing pull request may be broken,
so it is desiable to merge the base branch to the pull request and rerun CI.
In case many developers contribute to the repository and many pull requests are open at the same time,
the task `merge the base branch to the pull request and rerun CI` is bothersome and harm the Developer Productivity.

`run-ci` solves the problem by running CI and merging the base branch in CI when the base branch is updated.

`run-ci` provides the following sub commands.

* update-pr: run CI by updating the reference of the pull request's branch for a moment and restoring it immediately

`update-pr` does the following things by GitHub API

* [list pull requests](https://docs.github.com/en/rest/reference/pulls#list-pull-requests)
* [push an empty commit](https://docs.github.com/en/rest/reference/git#create-a-commit). We can skip the CI of this commit by commit message in many CI services
* [update the pull request's reference to the empty commit](https://docs.github.com/en/rest/reference/git#update-a-reference)
* [restore the pull request's reference immediately](https://docs.github.com/en/rest/reference/git#update-a-reference)

## How to use

Use `run-ci` in CI.

* In CI of the base branch, run `run-ci update-pr` to run CI of pull requests
* In CI of the pull request, merge the base branch to the feature branch

## Example

Coming soon.

## Install

Download the binary from [GitHub Releases](https://github.com/suzuki-shunsuke/run-ci/releases).

## Supported services

run-ci supports only GitHub, so other services like GitLab and Bitbucket aren't supported.
run-ci doesn't depend on the API of CI services, so any CI services are supported.

## GitHub Access Token

GitHub Access Token is required to get pull requests and create empty commits by the GitHub API.
To create empty commits, the write permission is required.

## Usage

```
$ run-ci help
NAME:
   run-ci - run CI automatically when pull request's base branch is updated. https://github.com/suzuki-shunsuke/run-ci

USAGE:
   run-ci [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   update-pr  run pull requests' CI
   init       generate a configuration file if it doesn't exist
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

### init

Generate a configuration file.

```
$ run-ci init --help
NAME:
   run-ci init - generate a configuration file if it doesn't exist

USAGE:
   run-ci init [command options] [arguments...]

OPTIONS:
   --help, -h  show help (default: false)
```

```
$ run-ci init # .run-ci.yaml is generated.
```

### update-pr

Run pull requests' CI.

```
$ run-ci update-pr --help
NAME:
   run-ci update-pr - run pull requests' CI

USAGE:
   run-ci update-pr [command options] [arguments...]

OPTIONS:
   --owner value             repository owner
   --repo value              repository name
   --github-token value      GitHub Access Token [$GITHUB_TOKEN, $GITHUB_ACCESS_TOKEN]
   --base value              base branch. Either the option 'base' or 'all' should be set
   --all                     get pull requests without specifying the base branch. Either the option 'base' or 'all' should be set (default: false)
   --config value, -c value  configuration file path
   --help, -h                show help (default: false)
```

Run pull requests' CI whose base branch is the master branch.

```
$ run-ci update-pr --base master
```

If you run CI regardless of the base branch, please specify `--all`.

```
$ run-ci update-pr --all
```

## Configuration

The configuration file path can be specified with the `--config (-c)` option.
If the confgiuration file path isn't specified, the file named `.run-ci.yml` or `.run-ci.yaml` would be searched from the current directory to the root directory.

```yaml
---
owner: suzuki-shunsuke # repository owner
repo: run-ci # repository name
empty_commit_msg: "[ci skip]" # empty commit's commit message
expr: "true" # expression to filter pull requests
github_token: xxx # GitHub Access Token
log_level: info # log level. To output the debug log, please set "debug"
```

### Expression

`run-ci update-pr` supports the filtering of pull requests by the expression.
After listing pull requests, the expression is evaluated per pull request.
And `run-ci` runs CI of only pull requests whose evaluation result is true.

ex.

```yaml
# .run-ci.yaml
# exclude pull requests which have the label "ignore-run-ci" or the author is "renovate[bot]"
expr: |
  "ignore-run-ci" not in util.labelNames(pr.labels) && pr.user.login != "renovate[bot]"
```

As the expression engine, [antonmedv/expr](https://github.com/antonmedv/expr) is used.
About the language, please see [Language Definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md).

The language is simple but very powerful.
We can express the condition of the pull request flexibility.

To understand the language, we recommend to write a simple Go code with [antonmedv/expr](https://github.com/antonmedv/expr).
We can try  [antonmedv/expr](https://github.com/antonmedv/expr) with [The Go Playground](https://play.golang.org).

ex. https://play.golang.org/p/wZZnybcioX1

#### Expression variables

* `pr`: pull request. Please see the response body of [list pull requests API](https://docs.github.com/en/rest/reference/pulls#list-pull-requests) 
* `util`: The utility functions.
  * `labelNames(pr.labels) []string`: return the list of the pull request label names.
  * `env`: The function to get the value of the environment variable. If the environment variable isn't set, the empty string is returned. [os.Getenv](https://golang.org/pkg/os/#Getenv) is used

## LICENSE

[MIT](LICENSE)
