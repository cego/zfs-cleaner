package main

import (
	"fmt"
	"github.com/cego/zfs-cleaner/zfs"
	"os/exec"
	"strings"
)

type (
	todo struct {
		comment string
		command string
		args    []string
	}
)

func newDestroy(snapshot *zfs.Snapshot) todo {
	return todo{
		comment: fmt.Sprintf("Destroying %s (Age %s)", snapshot.Name, now.Sub(snapshot.Creation)),
		command: commandName,
		args:    []string{"destroy", snapshot.Name},
	}
}

func newComment(format string, args ...interface{}) todo {
	return todo{
		comment: fmt.Sprintf(format, args...),
		command: "",
		args:    nil,
	}
}

func (t *todo) Do() error {
	if verbose && t.comment != "" {
		fmt.Fprintf(stdout, "### %s\n", t.comment)
	}

	if (verbose || dryrun) && t.command != "" {
		fmt.Fprintf(stdout, "# Running '%s %s'\n", t.command, strings.Join(t.args, " "))
	}

	if !dryrun && t.command != "" {
		output, err := exec.Command(t.command, t.args...).Output() //nolint:gosec

		fmt.Fprintf(stdout, "%s", string(output))

		if err != nil {
			return err
		}
	}

	return nil
}
