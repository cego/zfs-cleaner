package zfs

import (
	"fmt"
	"os/exec"
	"strings"
)

type Executor interface {
	GetSnapshotList(dataset string) ([]byte, error)
	GetFilesystems() ([]byte, error)
	HasSnapshot(dataset string) (bool, error)
	HasHolds(dataset string) (bool, error)
	DestroySnapshot(dataset string) ([]byte, error)
}

var _ Executor = (*executorImpl)(nil)

type executorImpl struct {
	zfsCommandName string
}

func NewExecutor() Executor {
	return &executorImpl{
		zfsCommandName: "/sbin/zfs",
	}
}

func (z *executorImpl) GetSnapshotList(dataset string) ([]byte, error) {
	commandArguments := []string{"list", "-t", "snapshot", "-o", "name,creation", "-s", "creation", "-d", "1", "-H", "-p", "-r", dataset}
	output, err := exec.Command(z.zfsCommandName, commandArguments...).Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		return nil, fmt.Errorf("failed to get snapshot list for dataset: %s error: %s", dataset, exitError.Stderr)
	}
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (z *executorImpl) GetFilesystems() ([]byte, error) {
	commandArguments := []string{"list", "-t", "filesystem", "-o", "name", "-H"}
	output, err := exec.Command(z.zfsCommandName, commandArguments...).Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		return nil, fmt.Errorf("failed to get filesystem list error: %s", exitError.Stderr)
	}
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (z *executorImpl) HasSnapshot(dataset string) (bool, error) {
	argsStr := fmt.Sprintf("list -t snapshot -o name %s -H -d 1", dataset)
	args := strings.Fields(argsStr)
	output, err := exec.Command(z.zfsCommandName, args...).Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		return false, fmt.Errorf("failed to get snapshot list to see if it has snapshots for dataset: %s error: %s", dataset, exitError.Stderr)
	}
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}

func (z *executorImpl) HasHolds(snapshot string) (bool, error) {
	argsStr := fmt.Sprintf("holds -H %s", snapshot)
	args := strings.Fields(argsStr)
	output, err := exec.Command(z.zfsCommandName, args...).Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		return false, fmt.Errorf("failed to get snapshot holds to see if the snapshots has any holds: %s error: %s", snapshot, exitError.Stderr)
	}
	if err != nil {
		return false, err
	}
	return len(output) != 0, nil
}

func (z *executorImpl) DestroySnapshot(snapshot string) ([]byte, error) {
	output, err := exec.Command(z.zfsCommandName, "destroy", snapshot).Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		return output, fmt.Errorf("failed to destroy snapshot: %s error: %s", snapshot, exitError.Stderr)
	}
	if err != nil {
		return nil, err
	}
	return output, nil
}
