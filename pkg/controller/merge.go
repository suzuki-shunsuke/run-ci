package controller

import "context"

func (ctrl Controller) Merge(ctx context.Context) error {
	if err := ctrl.Git.Fetch(ctx, "master"); err != nil {
		return err
	}
	if err := ctrl.Git.Merge(ctx, "origin/master"); err != nil {
		return err
	}
	return nil
}
