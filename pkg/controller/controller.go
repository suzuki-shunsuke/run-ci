package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/run-ci/pkg/expr"
	gh "github.com/suzuki-shunsuke/run-ci/pkg/github"
)

type ParamsUpdatePR struct {
	Branch         string
	EmptyCommitMsg string
	Logger         *logrus.Entry
}

func (ctrl Controller) UpdatePR(ctx context.Context) error { //nolint:funlen
	prs, _, err := ctrl.GitHub.ListPRs(ctx, gh.ParamsListPRs{
		Owner: ctrl.Config.Owner,
		Repo:  ctrl.Config.Repo,
		Base:  ctrl.Config.Base,
	})
	if err != nil {
		return err
	}
	if len(prs) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	for _, pr := range prs {
		logger := logrus.WithFields(logrus.Fields{
			"owner":     ctrl.Config.Owner,
			"repo":      ctrl.Config.Repo,
			"pr_number": *pr.Number,
			"head_ref":  *pr.Head.Ref,
		})
		logger.Debug("start to proceed the pull request")
		b, err := json.Marshal(pr)
		if err != nil {
			logger.WithError(err).Error("failed to marshal the request body")
			continue
		}
		m := map[string]interface{}{}
		if err := json.Unmarshal(b, &m); err != nil {
			logger.WithError(err).Error("failed to unmarshal the request body")
			continue
		}
		M := map[string]interface{}{
			"pr": m,
			"util": map[string]interface{}{
				"env":        os.Getenv,
				"labelNames": expr.LabelNames,
			},
		}
		matched, err := ctrl.Expr.Match(M)
		if err != nil {
			logger.WithError(err).Error("failed to evaluate whether the pull request matches the condition")
			continue
		}
		if !matched {
			logger.Debug("this pull request is skipped because it doesn't match the condition")
			continue
		}

		logger.Info("update the pull request")
		wg.Add(1)
		go func(pr *github.PullRequest, logger *logrus.Entry) {
			defer wg.Done()
			if err := ctrl.updatePR(ctx, ParamsUpdatePR{
				Branch:         *pr.Head.Ref,
				EmptyCommitMsg: ctrl.Config.EmptyCommitMsg,
				Logger:         logger,
			}); err != nil {
				logger.WithError(err).Error("failed to update the pull request")
			}
		}(pr, logger)
	}
	wg.Wait()
	return nil
}

func (ctrl Controller) updatePR(ctx context.Context, params ParamsUpdatePR) error {
	if ctrl.Config.GitCommand.Use {
		return ctrl.updatePRByGit(ctx, params)
	}
	return ctrl.updatePRByAPI(ctx, params)
}

func (ctrl Controller) updatePRByGit(ctx context.Context, params ParamsUpdatePR) error {
	branch := params.Branch
	emptyCommitMsg := params.EmptyCommitMsg
	if err := ctrl.Git.Fetch(ctx, "origin", branch); err != nil {
		return err
	}
	if err := ctrl.Git.Checkout(ctx, branch); err != nil {
		return err
	}
	if err := ctrl.Git.CommitEmpty(ctx, emptyCommitMsg); err != nil {
		return err
	}
	if err := ctrl.Git.Push(ctx, branch); err != nil {
		return err
	}
	if err := ctrl.Git.Reset(ctx); err != nil {
		return err
	}
	if err := ctrl.Git.PushForce(ctx, branch); err != nil {
		return err
	}
	return nil
}

func (ctrl Controller) updatePRByAPI(ctx context.Context, params ParamsUpdatePR) error { //nolint:funlen
	branch := "heads/" + params.Branch
	emptyCommitMsg := params.EmptyCommitMsg
	logger := params.Logger
	if logger == nil {
		logger = logrus.WithFields(logrus.Fields{
			"owner": ctrl.Config.Owner,
			"repo":  ctrl.Config.Repo,
		})
	}

	ref, _, err := ctrl.GitHub.GetRef(ctx, gh.ParamsGetRef{
		Owner: ctrl.Config.Owner,
		Repo:  ctrl.Config.Repo,
		Ref:   branch,
	})
	if err != nil {
		return fmt.Errorf("failed to get a Git Reference by GitHub API %s/%s %s: %w", ctrl.Config.Owner, ctrl.Config.Repo, branch, err)
	}

	sha := *ref.Object.SHA

	cmt, _, err := ctrl.GitHub.GetCommit(ctx, gh.ParamsGetCommit{
		Owner: ctrl.Config.Owner,
		Repo:  ctrl.Config.Repo,
		SHA:   sha,
	})
	if err != nil {
		return fmt.Errorf("failed to get a commit by GitHub API %s/%s %s: %w", ctrl.Config.Owner, ctrl.Config.Repo, sha, err)
	}

	newCmt, _, err := ctrl.GitHub.CreateEmptyCommit(ctx, gh.ParamsCreateEmptyCommit{
		Owner:     ctrl.Config.Owner,
		Repo:      ctrl.Config.Repo,
		CommitMsg: emptyCommitMsg,
		Parent:    cmt,
	})
	if err != nil {
		return fmt.Errorf("failed to create an empty commit by GitHub API %s/%s %s: %w", ctrl.Config.Owner, ctrl.Config.Repo, *cmt.Tree.SHA, err)
	}
	logger.WithFields(logrus.Fields{
		"sha": *newCmt.SHA,
	}).Debug("an empty commit is created by GitHub API")

	_, _, err = ctrl.GitHub.UpdateRef(ctx, gh.ParamsUpdateRef{
		Owner: ctrl.Config.Owner,
		Repo:  ctrl.Config.Repo,
		Ref:   branch,
		SHA:   *newCmt.SHA,
	})
	if err != nil {
		return fmt.Errorf("failed to update a git reference by GitHub API %s/%s %s %s: %w", ctrl.Config.Owner, ctrl.Config.Repo, branch, sha, err)
	}
	logger.WithFields(logrus.Fields{
		"sha": *newCmt.SHA,
		"ref": branch,
	}).Debug("the branch reference is changed to the empty commit by GitHub API")

	interval := 5 * time.Second //nolint:gomnd
	logger.WithFields(logrus.Fields{
		"interval": interval,
	}).Debug("wait until the pull request's tracking branch is synchronized with the source branch for the pull request")
	// https://github.community/t/what-is-a-pull-request-synchronize-event/14784/2
	// > A pull_request event it’s only triggered when the pull request’s tracking branch is synchronized with the source branch for the pull request,
	// > and that happens when the source branch is updated.
	// wait for GitHub to synchronize the change of the source branch.
	// Otherwise, the `pull_request` event doesn't occur and the webhook isn't sent to the CI service.
	timer := time.NewTimer(interval)
	select {
	case <-timer.C:
	case <-ctx.Done():
	}

	_, _, err = ctrl.GitHub.UpdateRef(ctx, gh.ParamsUpdateRef{
		Owner: ctrl.Config.Owner,
		Repo:  ctrl.Config.Repo,
		Ref:   branch,
		SHA:   sha,
	})
	if err != nil {
		return fmt.Errorf("failed to update a git reference by GitHub API %s/%s %s %s: %w", ctrl.Config.Owner, ctrl.Config.Repo, branch, sha, err)
	}
	logger.WithFields(logrus.Fields{
		"sha": sha,
		"ref": branch,
	}).Debug("the branch reference is rollback by GitHub API")

	return nil
}
