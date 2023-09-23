package saboteur

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Merge(ctx context.Context, authHeaderFunc func(ctx context.Context) (string, error), owner, repo string, mr MergeableMR) (err error) {
	tmpDir, err := os.MkdirTemp("", "git-clone-*")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}

	defer func() {
		if cleanupErr := os.RemoveAll(tmpDir); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	token, err := authHeaderFunc(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving auth header: %w", err)
	}

	if out, err := stripEnv(exec.CommandContext(
		ctx,
		"git", "clone",
		"-c", "http.extraHeader=Authorization: "+token,
		"--bare", "--depth=1", "--filter=blob:none",
		"-b", strings.TrimPrefix(mr.BaseRef, "refs/heads/"),
		fmt.Sprintf("https://github.com/%s/%s.git", owner, repo),
		tmpDir,
	)).CombinedOutput(); err != nil {
		return fmt.Errorf("error cloning: %s", out)
	}

	if out, err := stripEnv(exec.CommandContext(
		ctx,
		"git", "push",
		"origin",
		mr.Head+":"+mr.BaseRef,
	), "GIT_DIR="+tmpDir).CombinedOutput(); err != nil {
		return fmt.Errorf("error pushing: %s", out)
	}

	return nil
}

func stripEnv(cmd *exec.Cmd, additionalEnvVars ...string) *exec.Cmd {
	if cmd.Env == nil {
		cmd.Env = []string{
			"PATH=" + os.Getenv("PATH"),
		}
	}

	cmd.Env = append(cmd.Env, additionalEnvVars...)

	return cmd
}
