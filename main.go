package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cego/zfs-cleaner/conf"
	"github.com/cego/zfs-cleaner/zfs"
	"github.com/spf13/cobra"
)

var (
	verbose = false
	dryrun  = false

	commandName      = "zfs"
	commandArguments = []string{"list", "-t", "snapshot", "-o", "name,creation", "-s", "creation", "-r", "-H", "-p"}

	// This can be set to a specific time for testing.
	now = time.Now()

	// This can be changed to true when testing.
	panicBail = false

	rootCmd = &cobra.Command{
		Use:   "zfs-cleaner [config file]",
		Short: "Tool for destroying ZFS snapshots after predefined retention periods",
		RunE:  clean,
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryrun, "dryrun", "n", false, "Do nothing destructive, only print")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Be more verbose")
}

func getList(name string) (zfs.SnapshotList, error) {
	// Output could be cached.
	output, err := exec.Command(commandName, commandArguments...).Output()

	if err != nil {
		return nil, err
	}

	return zfs.NewSnapshotListFromOutput(output, name)
}

func readConf(path string) (*conf.Config, error) {
	conf := &conf.Config{}

	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %s", path, err.Error())
	}

	err = conf.Read(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %s", path, err.Error())
	}

	return conf, nil
}

func processAll(now time.Time, conf *conf.Config) ([]zfs.SnapshotList, error) {
	lists := []zfs.SnapshotList{}
	for _, plan := range conf.Plans {
		for _, path := range plan.Paths {
			list, err := getList(path)
			if err != nil {
				return nil, err
			}

			list.KeepLatest(plan.Latest)
			for _, period := range plan.Periods {
				start := now.Add(-period.Age)

				list.Sieve(start, period.Frequency)
			}

			lists = append(lists, list)
		}
	}

	return lists, nil
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		if panicBail {
			panic(err.Error())
		}

		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func clean(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("%s /path/to/config.conf", cmd.Name())
	}

	conf, err := readConf(args[0])
	if err != nil {
		return err
	}

	lists, err := processAll(now, conf)
	if err != nil {
		return err
	}

	// Start by generating a list of stuff to do.
	todos := []todo{}

	for _, list := range lists {
		for _, snapshot := range list {
			if !snapshot.Keep {
				todos = append(todos, newDestroy(snapshot.Name))
			} else {
				todos = append(todos, newComment("Keep %s (Age %s)", snapshot.Name, now.Sub(snapshot.Creation)))
			}
		}
	}

	// And then do it! :-)
	for _, todo := range todos {
		err := todo.Do()
		if err != nil {
			return err
		}
	}

	return nil
}
