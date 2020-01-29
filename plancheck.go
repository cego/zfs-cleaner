package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cego/zfs-cleaner/conf"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(plancheckCmd)
}

var plancheckCmd = &cobra.Command{
	Use:   "plancheck [config file]",
	Short: "Print mounts that have no plan and exit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("%s /path/to/config.conf", cmd.Name())
		}

		confFile, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("failed to open %s: %s", args[0], err.Error())
		}
		defer confFile.Close()

		conf, err := readConf(confFile)
		if err != nil {
			return err
		}
		return planCheck(conf)
	},
}

func planCheck(conf *conf.Config) error {
	args := []string{"list", "-t", "filesystem", "-o", "name", "-H"}

	output, err := exec.Command(commandName, args...).Output()
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
			fmt.Printf("No plan found for path: '%s'\n", store)
		}
	}

	return nil
}
