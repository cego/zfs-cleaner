package main

import (
	"errors"
	"github.com/cego/zfs-cleaner/zfs"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cego/zfs-cleaner/conf"
)

func init() {
	panicBail = true

	// Mute normal output when running tests.
	stdout = ioutil.Discard
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}

func TestReadConf(t *testing.T) {
	expected := &conf.Config{
		Plans: []conf.Plan{
			{
				Name:   "buh",
				Paths:  []string{"/buh"},
				Latest: 10,
				Periods: []conf.Period{
					{
						Frequency: 24 * time.Hour,
						Age:       30 * 24 * time.Hour,
					},
				},
			},
		},
	}

	content := []byte(`
plan buh {
path /buh
keep 1d for 30d
keep latest 10
}
`)
	tmpfile, err := ioutil.TempFile("/dev/shm", "test.TestReadConfSyntaxError")
	if err != nil {
		t.Fatalf("Failed to create config file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	if err != nil {
		t.Fatalf("Failed to write config file: %s", err.Error())
	}

	defer tmpfile.Close()
	_, _ = tmpfile.Seek(0, 0)

	conf, err := readConfig(tmpfile)
	if err != nil {
		t.Errorf("readConfig() failed: %s", err.Error())
	}

	if conf == nil {
		t.Errorf("readConfig() did not return a config")
	}

	if !reflect.DeepEqual(conf, expected) {
		t.Errorf("readConfig() did not return expected config")
	}
}

func TestReadConfNoFile(t *testing.T) {
	err := clean(nil, []string{"/non/existing.conf"})
	if err == nil {
		t.Errorf("Failed to error on non-existing config file")
	}
}

func TestReadConfSyntaxError(t *testing.T) {
	content := []byte("syntax error")
	tmpfile, err := ioutil.TempFile("/dev/shm", "test.TestReadConfSyntaxError")
	if err != nil {
		t.Fatalf("Failed to create config file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	if err != nil {
		t.Fatalf("Failed to write config file: %s", err.Error())
	}

	defer tmpfile.Close()
	_, _ = tmpfile.Seek(0, 0)

	conf, err := readConfig(tmpfile)
	if err == nil {
		t.Errorf("Failed to return error for broken config file")
	}

	if conf != nil {
		t.Errorf("Returned non-nil config for broken config file")
	}
}

var _ zfs.Executor = (*testExecutor)(nil)

type testExecutor struct {
	zfsCommandName        string
	getSnapshotListResult []byte
	getSnapshotListError  error
}

func (t *testExecutor) GetSnapshotList(dataset string) ([]byte, error) {
	return t.getSnapshotListResult, t.getSnapshotListError
}

func (t *testExecutor) GetFilesystems() ([]byte, error) {
	panic("implement me")
}

func (t *testExecutor) HasSnapshot(dataset string) (bool, error) {
	panic("implement me")
}

func (t *testExecutor) DestroySnapshot(dataset string) ([]byte, error) {
	return nil, nil
}

func TestProcessAll(t *testing.T) {
	zfsTestExecutor := testExecutor{
		getSnapshotListResult: []byte(`playground/fs1@snap1	1492989570
playground/fs1@snap2	1492989572
playground/fs1@snap3	1492989573
playground/fs1@snap4	1492989574
playground/fs1@snap5	1492989587
`),
	}

	conf := &conf.Config{
		Plans: []conf.Plan{
			{
				Name:   "buh",
				Paths:  []string{"playground/fs1"},
				Latest: 10,
				Periods: []conf.Period{
					{
						Frequency: 24 * time.Hour,
						Age:       30 * 24 * time.Hour,
					},
				},
			},
		},
	}

	lists, err := processAll(time.Now(), conf, &zfsTestExecutor)
	if err != nil {
		t.Errorf("processAll() returned error: %s", err.Error())
	}

	if lists == nil {
		t.Fatalf("processAll() returned nil lists")
	}

	if len(lists) != 1 {
		t.Errorf("processAll() returned wrong number of lists, got %d", len(lists))
	}
}

func TestProcessAllFail(t *testing.T) {
	zfsTestExecutor := testExecutor{
		getSnapshotListError: errors.New("test fail"),
	}

	conf := &conf.Config{
		Plans: []conf.Plan{
			{
				Name:   "buh",
				Paths:  []string{"playground/fs1"},
				Latest: 10,
				Periods: []conf.Period{
					{
						Frequency: 24 * time.Hour,
						Age:       30 * 24 * time.Hour,
					},
				},
			},
		},
	}

	lists, err := processAll(time.Now(), conf, &zfsTestExecutor)
	if err == nil {
		t.Errorf("processAll() did not return error")
	}

	if lists != nil {
		t.Errorf("processAll() returned lists")
	}
}

func TestMainNoArguments(t *testing.T) {
	os.Args = []string{os.Args[0]}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic for no arguments")
		}
	}()

	main()
}

func TestMainNoConfig(t *testing.T) {
	os.Args = []string{os.Args[0], "/non-existing-config.conf"}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic for no arguments")
		}
	}()

	main()
}

func TestMainNoZFS(t *testing.T) {
	content := []byte(`
plan buh {
path /buh
keep 1d for 30d
keep latest 10
}
`)
	tmpfile, err := ioutil.TempFile("/dev/shm", "test.TestReadConfSyntaxError")
	if err != nil {
		t.Fatalf("Failed to create config file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	if err != nil {
		t.Fatalf("Failed to write config file: %s", err.Error())
	}

	err = tmpfile.Close()
	if err != nil {
		t.Fatalf("Failed to close config file: %s", err.Error())
	}

	os.Args = []string{os.Args[0], tmpfile.Name()}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic for no arguments")
		}
	}()
	main()
}

func TestMainFull(t *testing.T) {
	now = time.Unix(1492993419, 0)
	verbose = true
	content := []byte(`
plan buh {
path playground/fs1
keep 60s for 30d
keep latest 1
}
`)
	tmpfile, err := ioutil.TempFile("/dev/shm", "test.TestReadConfSyntaxError")
	if err != nil {
		t.Fatalf("Failed to create config file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	if err != nil {
		t.Fatalf("Failed to write config file: %s", err.Error())
	}

	err = tmpfile.Close()
	if err != nil {
		t.Fatalf("Failed to close config file: %s", err.Error())
	}

	os.Args = []string{os.Args[0], tmpfile.Name()}

	zfsExecutor = &testExecutor{
		getSnapshotListResult: []byte(`playground/fs1@snap1	1492989570
playground/fs1@snap2	1492989572
playground/fs1@snap3	1492989573
playground/fs1@snap4	1492989574
playground/fs1@snap5	1492989587
`),
	}
	main()
}

func TestConcurrency(t *testing.T) {
	var lock sync.Mutex

	tmpfile, _ := ioutil.TempFile("/dev/shm", "test.TestReadConfSyntaxError")
	defer os.Remove(tmpfile.Name())

	// An empty configuration file is valid too.
	_, _ = tmpfile.Write([]byte(""))
	defer tmpfile.Close()

	// This will force clean() to wait for our mainWaitGroup.Done().
	mainWaitGroup.Add(1)

	lock.Lock()

	var cleanErr error
	go func() {
		cleanErr = clean(nil, []string{tmpfile.Name()})

		lock.Unlock()
	}()

	// Give some time for the first clean() to acquire the lock.
	time.Sleep(time.Millisecond * 100)

	err := clean(nil, []string{tmpfile.Name()})
	if err == nil {
		t.Fatalf("clean() failed to detect locked configuration file")
	}

	// Let the first clean() exit.
	mainWaitGroup.Done()

	lock.Lock()

	if cleanErr != nil {
		t.Fatalf("clean() returned an error: %s", cleanErr.Error())
	}
}
