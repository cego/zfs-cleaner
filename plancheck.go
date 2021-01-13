package main

import (
	"fmt"
	"github.com/cego/zfs-cleaner/zfs"
	"os"
	"strings"

	"github.com/cego/zfs-cleaner/conf"
	"github.com/spf13/cobra"
)

func AddPlanCheckCommand(zfsExecutor zfs.Executor) {
	ignoreEmpty := false
	planCheckCmd := &cobra.Command{
		Use:   "plancheck [config file]",
		Short: "Print mounts that have no plan and exit",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("%s /path/to/config.config", cmd.Name())
			}
			configFile, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open %s: %s", args[0], err.Error())
			}
			defer configFile.Close()
			config, err := readConfig(configFile)
			if err != nil {
				return err
			}
			return planCheck(zfsExecutor, config, ignoreEmpty)
		},
	}
	planCheckCmd.PersistentFlags().BoolVar(&ignoreEmpty, "ignore-empty", false, "Ignore file systems with no snapshots")
	rootCmd.AddCommand(planCheckCmd)
}

func hasSnapshots(zfsExecutor zfs.Executor, dataset string) bool {
	hasSnapshot, err := zfsExecutor.HasSnapshot(dataset)
	if err != nil {
		return false
	}
	return hasSnapshot
}

func planCheck(zfsExecutor zfs.Executor, conf *conf.Config, ignoreEmpty bool) error {
	output, err := zfsExecutor.GetFilesystems()
	if err != nil {
		return err
	}
	m := map[string]bool{}
	for _, plan := range conf.Plans {
		for _, path := range plan.Paths {
			m[path] = true
		}
	}
	for _, store := range strings.Fields(string(output)) {
		if !m[store] {
			if ignoreEmpty && !hasSnapshots(zfsExecutor, store) {
				continue
			}
			fmt.Printf("No plan found for path: '%s'\n", store)
		}
	}
	return nil
}
