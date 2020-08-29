package controller

import "context"

func (ctrl Controller) Merge(ctx context.Context, remote, branch string) error {
	if err := ctrl.Git.Fetch(ctx, remote, branch); err != nil {
		return err
	}
	if err := ctrl.Git.Merge(ctx, remote+"/"+branch); err != nil {
		return err
	}
	return nil
}
