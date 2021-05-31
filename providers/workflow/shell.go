package workflow

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/workflow/runtime"
	"os/exec"
)

func runBash(s string, ctx *runtime.WorkflowContext) ([]byte, error) {
	path, err := exec.LookPath("bash")
	if err != nil {
		return runShell(s, ctx)
	}

	cmd := exec.Command(path, "-c", s)
	cmd.Env = ctx.EnvStrings()

	return cmd.Output()
}

func runShell(s string, ctx *runtime.WorkflowContext) ([]byte, error) {
	path, err := exec.LookPath("sh")
	if err != nil {
		return nil, errors.Wrap(err, "unable to run step")
	}

	cmd := exec.Command(path, "-c", s)
	cmd.Env = ctx.EnvStrings()

	return cmd.Output()
}

func runCmd(s string, ctx *runtime.WorkflowContext) ([]byte, error) {
	path, err := exec.LookPath("cmd")
	if err != nil {
		return nil, fmt.Errorf("cmd not found")
	}
	cmd := &exec.Cmd{
		Path: path,
		Args: []string{"/C", s},
		Env:  ctx.EnvStrings(),
	}

	return cmd.Output()
}
