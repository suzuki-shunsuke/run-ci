package cli

import (
	"io/ioutil"
	"os"

	"github.com/urfave/cli/v2"
)

const cfgTpl = `---
# # Configuration file of run-ci, which is a CLI tool to run CI.
# # https://github.com/suzuki-shunsuke/run-ci
# owner: ""
# repo: ""
# expr: "true"
# empty_commit_msg: "[ci skip]"
# github_token: ""
# git_command:
#   use: false
#   user_name: run-ci
#   user_email: run-ci@example.com`

func (runner Runner) initAction(c *cli.Context) error {
	if _, err := os.Stat(".run-ci.yml"); err == nil {
		return nil
	}
	if _, err := os.Stat(".run-ci.yaml"); err == nil {
		return nil
	}
	if err := ioutil.WriteFile(".run-ci.yaml", []byte(cfgTpl), 0o755); err != nil { //nolint:gosec
		return err
	}
	return nil
}
