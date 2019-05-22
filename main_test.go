package main

import (
	"io/ioutil"
	"os"
	"reflect"
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

func TestGetList(t *testing.T) {
	// We override the zfs command in tests, to avoid running the real zfs
	// binary.
	commandName = "echo"
	commandArguments = []string{"-e", "-n", "playground/fs1@snap1\t1492989570\nplayground/fs1@snap2\t1492989572\nplayground/fs1@snap3\t1492989573\nplayground/fs1@snap4\t1492989574\nplayground/fs1@snap5\t1492989587\n"}

	list, err := getList("playground/fs1")
	if err != nil {
		t.Fatalf("getList() returned error: %s", err.Error())
	}

	if len(list) != 5 {
		t.Fatalf("getList() returned wrong number of snapshots, got %d, expected 5", len(list))
	}
}

func TestGetListMissingBinary(t *testing.T) {
	commandName = "/non/existing-zfs-command"

	list, err := getList("/pool/fs1")
	if err == nil {
		t.Fatalf("getList() failed to error on non-existing ZFS binary")
	}

	if list != nil {
		t.Fatalf("getList() non-nil list for error")
	}
}

func TestGetListError(t *testing.T) {
	commandName = "false"

	list, err := getList("/pool/fs1")
	if err == nil {
		t.Fatalf("getList() failed to error on ZFS binary returning 1")
	}

	if list != nil {
		t.Fatalf("getList() non-nil list for error")
	}
}

func TestReadConf(t *testing.T) {
	expected := &conf.Config{
		Plans: []conf.Plan{
			conf.Plan{
				Name:   "buh",
				Paths:  []string{"/buh"},
				Latest: 10,
				Periods: []conf.Period{
					conf.Period{
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
	tmpfile.Seek(0, 0)

	conf, err := readConf(tmpfile)
	if err != nil {
		t.Errorf("readConf() failed: %s", err.Error())
	}

	if conf == nil {
		t.Errorf("readConf() did not return a config")
	}

	if !reflect.DeepEqual(conf, expected) {
		t.Errorf("readConf() did not return expected config")
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
	tmpfile.Seek(0, 0)

	conf, err := readConf(tmpfile)
	if err == nil {
		t.Errorf("Failed to return error for broken config file")
	}

	if conf != nil {
		t.Errorf("Returned non-nil config for broken config file")
	}
}

func TestProcessAll(t *testing.T) {
	commandName = "echo"
	commandArguments = []string{"-e", "-n", "playground/fs1@snap1\t1492989570\nplayground/fs1@snap2\t1492989572\nplayground/fs1@snap3\t1492989573\nplayground/fs1@snap4\t1492989574\nplayground/fs1@snap5\t1492989587\n"}

	conf := &conf.Config{
		Plans: []conf.Plan{
			conf.Plan{
				Name:   "buh",
				Paths:  []string{"playground/fs1"},
				Latest: 10,
				Periods: []conf.Period{
					conf.Period{
						Frequency: 24 * time.Hour,
						Age:       30 * 24 * time.Hour,
					},
				},
			},
		},
	}

	lists, err := processAll(time.Now(), conf)
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
	commandName = "false"
	commandArguments = []string{"-e", "-n", "playground/fs1@snap1\t1492989570\nplayground/fs1@snap2\t1492989572\nplayground/fs1@snap3\t1492989573\nplayground/fs1@snap4\t1492989574\nplayground/fs1@snap5\t1492989587\n"}

	conf := &conf.Config{
		Plans: []conf.Plan{
			conf.Plan{
				Name:   "buh",
				Paths:  []string{"playground/fs1"},
				Latest: 10,
				Periods: []conf.Period{
					conf.Period{
						Frequency: 24 * time.Hour,
						Age:       30 * 24 * time.Hour,
					},
				},
			},
		},
	}

	lists, err := processAll(time.Now(), conf)
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
	commandName = "false"

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
	commandName = "echo"
	commandArguments = []string{"-e", "-n", "playground/fs1@snap1\t1492989570\nplayground/fs1@snap2\t1492989572\nplayground/fs1@snap3\t1492989573\nplayground/fs1@snap4\t1492989574\nplayground/fs1@snap5\t1492989587\n"}
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

	main()
}

func TestConcurrency(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("/dev/shm", "test.TestReadConfSyntaxError")
	defer os.Remove(tmpfile.Name())

	// An empty configuration file is valid too.
	_, _ = tmpfile.Write([]byte(""))
	defer tmpfile.Close()

	// This will force clean() to wait for our mainWaitGroup.Done().
	mainWaitGroup.Add(1)

	go func() {
		err := clean(nil, []string{tmpfile.Name()})
		if err != nil {
			t.Fatalf("clean() returned an error: %s", err.Error())
		}
	}()

	// Give some time for the first clean() to aquire the lock.
	time.Sleep(time.Millisecond * 100)

	err := clean(nil, []string{tmpfile.Name()})
	if err == nil {
		t.Fatalf("clean() failed to detect locked configuration file")
	}

	// Let the first clean() exit.
	mainWaitGroup.Done()
}
