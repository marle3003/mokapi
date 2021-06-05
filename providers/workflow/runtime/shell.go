package runtime

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
)

func runBash(s string, ctx *WorkflowContext) ([]byte, error) {
	path, err := exec.LookPath("bash")
	if err != nil {
		return runShell(s, ctx)
	}

	cmd := exec.Command(path, "-c", s)
	cmd.Env = ctx.Environ()

	return cmd.Output()
}

func runShell(s string, ctx *WorkflowContext) ([]byte, error) {
	path, err := exec.LookPath("sh")
	if err != nil {
		return nil, errors.Wrap(err, "unable to run step")
	}

	cmd := exec.Command(path, "-c", s)
	cmd.Env = ctx.Environ()

	return cmd.Output()
}

func runCmd(s string, ctx *WorkflowContext) ([]byte, error) {
	path, err := exec.LookPath("cmd")
	if err != nil {
		return nil, fmt.Errorf("cmd not found")
	}
	cmd := &exec.Cmd{
		Path: path,
		Args: []string{"/C", s},
		Env:  ctx.Environ(),
	}

	return cmd.Output()
}

func runPowershell(s string, ctx *WorkflowContext) ([]byte, error) {
	path, err := exec.LookPath("powershell.exe")

	if err != nil {
		return nil, fmt.Errorf("powershell not found")
	}

	//code := strings.ReplaceAll(s, "\"", "\\\"")
	code := fmt.Sprintf("& {%v}", s)
	args := []string{"-NoProfile", "-NonInteractive", "-Command", code}

	cmd := &exec.Cmd{
		Path: path,
		Args: args,
		Env:  ctx.Environ(),
	}

	return cmd.Output()
}
