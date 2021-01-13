package main

import (
	"fmt"
	"github.com/cego/zfs-cleaner/zfs"
)

type todo interface {
	Do() error
}

var (
	_ todo = (*destroySnapshot)(nil)
	_ todo = (*noop)(nil)
)

type destroySnapshot struct {
	comment     string
	zfsExecutor zfs.Executor
	snapshot    *zfs.Snapshot
}

type noop struct {
	comment string
}

func newDestroy(zfsExecutor zfs.Executor, snapshot *zfs.Snapshot) todo {
	return &destroySnapshot{
		comment:     fmt.Sprintf("Destroying %s (Age %s)", snapshot.Name, now.Sub(snapshot.Creation)),
		zfsExecutor: zfsExecutor,
		snapshot:    snapshot,
	}
}

func (d *destroySnapshot) Do() error {
	if verbose {
		_, _ = fmt.Fprintf(stdout, "### %s\n", d.comment)
	}
	if verbose || dryrun {
		_, _ = fmt.Fprintf(stdout, "# Running 'zfs destroy %s'\n", d.snapshot.Name)
	}
	if !dryrun {
		output, err := d.zfsExecutor.DestroySnapshot(d.snapshot.Name)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(stdout, "%s", string(output))
	}
	return nil
}

func newComment(format string, args ...interface{}) todo {
	return &noop{
		comment: fmt.Sprintf(format, args...),
	}
}

func (d *noop) Do() error {
	if verbose {
		_, _ = fmt.Fprintf(stdout, "### %s\n", d.comment)
	}
	return nil
}
