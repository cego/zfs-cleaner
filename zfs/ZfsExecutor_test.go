package zfs

import (
	"testing"
)

var _ Executor = (*TestExecutor)(nil)

type TestExecutor struct {
	zfsCommandName string
}

func (t *TestExecutor) GetSnapshotList(dataset string) ([]byte, error) {
	panic("implement me")
}

func (t *TestExecutor) GetFilesystems() ([]byte, error) {
	panic("implement me")
}

func (t *TestExecutor) HasSnapshot(dataset string) (bool, error) {
	panic("implement me")
}

func (t *TestExecutor) DestroySnapshot(dataset string) ([]byte, error) {
	panic("implement me")
}

func TestGetListMissingBinary(t *testing.T) {
	executor := TestExecutor{
		zfsCommandName: "/non/existing-zfs-command",
	}
	list, err := executor.GetSnapshotList("test")
	if err == nil {
		t.Fatalf("getList() failed to error on non-existing ZFS binary")
	}
	if list != nil {
		t.Fatalf("getList() non-nil list for error")
	}
}

func TestGetListError(t *testing.T) {
	executor := TestExecutor{
		zfsCommandName: "false",
	}

	list, err := executor.GetSnapshotList("test")
	if err == nil {
		t.Fatalf("getList() failed to error on ZFS binary returning 1")
	}

	if list != nil {
		t.Fatalf("getList() non-nil list for error")
	}
}
