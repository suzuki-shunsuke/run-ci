package cli

import (
	"errors"

	"github.com/suzuki-shunsuke/run-ci/pkg/controller"
	"github.com/suzuki-shunsuke/run-ci/pkg/execute"
	"github.com/suzuki-shunsuke/run-ci/pkg/git"
	"github.com/urfave/cli/v2"
)

var ErrInvalidArgsMerge = errors.New(`the number of arguments is invalid`)

func (runner Runner) mergeAction(c *cli.Context) error {
	cfg, err := runner.readConfig(c)
	if err != nil {
		return err
	}
	args := c.Args()
	if args.Len() != 1 {
		return ErrInvalidArgsMerge
	}

	ctrl := controller.Controller{
		Config: cfg,
		Git: git.New(git.ParamsNew{
			UserName:  cfg.GitCommand.UserName,
			UserEmail: cfg.GitCommand.UserEmail,
			Executor:  execute.New(),
		}),
	}

	return ctrl.Merge(c.Context, c.String("remote"), args.First())
}
