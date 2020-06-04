// +build linux darwin

package cmd

import (
	"context"
	osexec "os/exec"
	"syscall"

	"github.com/pkg/errors"
)

func exec(ctx context.Context, command string, args []string, env []string) error {
	argv0, err := osexec.LookPath(command)
	if err != nil {
		return errors.Wrap(err, "Error finding command")
	}

	argv := make([]string, 0, 1+len(args))
	argv = append(argv, command)
	argv = append(argv, args...)

	// Only return if the execution fails.
	return errors.Wrap(syscall.Exec(argv0, argv, env), "error executing command")
}
