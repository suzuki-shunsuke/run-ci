package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/suzuki-shunsuke/run-ci/pkg/expr"
	gh "github.com/suzuki-shunsuke/run-ci/pkg/github"
)

type ParamsUpdatePR struct {
	Branch         string
	EmptyCommitMsg string
}

func (ctrl Controller) UpdatePR(ctx context.Context) error {
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

	for _, pr := range prs {
		log.Println("start to proceed the pull request " + ctrl.Config.Owner + "/" + ctrl.Config.Repo + "/" + strconv.Itoa(*pr.Number) + " " + *pr.Head.Ref + " " + *pr.Title)
		b, err := json.Marshal(pr)
		if err != nil {
			log.Println(err)
			continue
		}
		m := map[string]interface{}{}
		if err := json.Unmarshal(b, &m); err != nil {
			log.Println(err)
			continue
		}
		M := map[string]interface{}{
			"pr":  m,
			"env": os.Getenv,
			"util": map[string]interface{}{
				"labelNames": expr.LabelNames,
			},
		}
		matched, err := ctrl.Expr.Match(M)
		if err != nil {
			log.Println(err)
			continue
		}
		if !matched {
			log.Println("this pull request is skipped because it doesn't match the condition")
			continue
		}

		log.Println("update the pull request " + ctrl.Config.Owner + "/" + ctrl.Config.Repo + "/" + strconv.Itoa(*pr.Number) + " " + *pr.Head.Ref + " " + *pr.Title)
		if err := ctrl.updatePR(ctx, ParamsUpdatePR{
			Branch:         *pr.Head.Ref,
			EmptyCommitMsg: ctrl.Config.EmptyCommitMsg,
		}); err != nil {
			log.Println(err)
			continue
		}
	}
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

func (ctrl Controller) updatePRByAPI(ctx context.Context, params ParamsUpdatePR) error {
	log.Println("update by GitHub API")
	branch := "heads/" + params.Branch
	emptyCommitMsg := params.EmptyCommitMsg

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

	_, _, err = ctrl.GitHub.UpdateRef(ctx, gh.ParamsUpdateRef{
		Owner: ctrl.Config.Owner,
		Repo:  ctrl.Config.Repo,
		Ref:   branch,
		SHA:   *newCmt.SHA,
	})
	if err != nil {
		return fmt.Errorf("failed to update a git reference by GitHub API %s/%s %s %s: %w", ctrl.Config.Owner, ctrl.Config.Repo, branch, sha, err)
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

	return nil
}
