package main

import (
	"fmt"
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

func newDestroy(name string) todo {
	return todo{
		comment: "Destroying " + name,
		command: commandName,
		args:    []string{"destroy", name},
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
		output, err := exec.Command(t.command, t.args...).Output()

		fmt.Fprintf(stdout, "%s", string(output))

		if err != nil {
			return err
		}
	}

	return nil
}
