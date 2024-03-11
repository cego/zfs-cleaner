package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/cego/zfs-cleaner/conf"
	"github.com/cego/zfs-cleaner/zfs"
	"github.com/spf13/cobra"
)

var (
	verbose     = false
	dryrun      = false
	showVersion = false
	// This can be set to a specific time for testing.
	now = time.Now()
	// tasks can be added to this for testing.
	mainWaitGroup sync.WaitGroup
	// This can be changed to true when testing.
	panicBail = false
	rootCmd   = &cobra.Command{
		Use:   "zfs-cleaner [config file]",
		Short: "Tool for destroying ZFS snapshots after predefined retention periods",
		RunE:  clean,
	}
	// Can be overridden when running tests.
	stdout      io.Writer = os.Stdout
	zfsExecutor zfs.Executor
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryrun, "dryrun", "n", false, "Do nothing destructive, only print")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Be more verbose")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "V", false, "Show version and exit")
	rootCmd.TraverseChildren = true
	zfsExecutor = zfs.NewExecutor()
}

func readConfig(r *os.File) (*conf.Config, error) {
	config := &conf.Config{}
	err := config.Read(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %s", r.Name(), err.Error())
	}
	return config, nil
}

func processAll(now time.Time, conf *conf.Config, zfsExecutor zfs.Executor) ([]zfs.SnapshotList, error) {
	lists := []zfs.SnapshotList{}
	for _, plan := range conf.Plans {
		for _, dataset := range plan.Paths {
			list := zfs.SnapshotList{}
			list, err := list.NewSnapshotListFromDataset(zfsExecutor, dataset)
			if err != nil {
				return nil, err
			}
			list.KeepNamed(plan.Protect)
			list.KeepLatest(plan.Latest)
			err = list.KeepHolds(zfsExecutor)
			if err != nil {
				return nil, err
			}
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
	AddPlanCheckCommand(zfsExecutor)
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
	if showVersion {
		printVersion()
	}
	if len(args) != 1 {
		return fmt.Errorf("%s /path/to/config.conf", cmd.Name())
	}
	configPath := args[0]
	confFile, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %s", args[0], err.Error())
	}
	defer confFile.Close()
	conf, err := readConfig(confFile)
	if err != nil {
		return err
	}
	fd := int(confFile.Fd())
	err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return fmt.Errorf("could not acquire lock on '%s'", confFile.Name())
	}
	// make sure to unlock :)
	defer func() {
		// We can ignore errors here, we're exiting anyway.
		_ = syscall.Flock(fd, syscall.LOCK_UN)
	}()
	lists, err := processAll(now, conf, zfsExecutor)
	if err != nil {
		return err
	}
	// Start by generating a list of stuff to do.
	todos := []todo{}
	// Print plan when verbose.
	if verbose {
		todos = append(todos, newComment("Config: '%s'", configPath))
		for _, plan := range conf.Plans {
			todos = append(todos, newComment("Plan: %+v", plan))
		}
	}
	for _, list := range lists {
		for _, snapshot := range list {
			if !snapshot.Keep {
				todos = append(todos, newDestroy(zfsExecutor, snapshot))
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
	mainWaitGroup.Wait()
	return nil
}
