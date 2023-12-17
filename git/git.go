package git

import (
	"context"
	"fmt"
	"l7-cloud-builder/cmd"
)

func Fetch(ctx context.Context, workdir string) error {
	var gitFetch = &cmd.Cmd{
		Command:     "git",
		CommandLine: "fetch",
		WorkingDir:  workdir,
	}

	if err := gitFetch.Run(ctx); err != nil {
		return fmt.Errorf("failed to update the repo: %w", err)
	}

	return nil
}

func CheckoutBranch(ctx context.Context, workdir, branch string) error {
	var gitCheckout = &cmd.Cmd{
		Command:      "git",
		CommandLine:  "checkout {branch}",
		WorkingDir:   workdir,
		Placeholders: map[string]string{"branch": branch},
	}

	if err := gitCheckout.Run(ctx); err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

func CheckoutTag(ctx context.Context, workdir, tag string) error {
	var gitCheckout = &cmd.Cmd{
		Command:      "git",
		CommandLine:  "checkout {tag}",
		WorkingDir:   workdir,
		Placeholders: map[string]string{"tag": tag},
	}

	if err := gitCheckout.Run(ctx); err != nil {
		return fmt.Errorf("failed to checkout tag: %w", err)
	}

	return nil
}

func Pull(ctx context.Context, workdir string) error {
	var gitPull = &cmd.Cmd{
		Command:     "git",
		CommandLine: "pull",
		WorkingDir:  workdir,
	}

	if err := gitPull.Run(ctx); err != nil {
		return fmt.Errorf("failed to pull the repo: %w", err)
	}

	return nil
}
