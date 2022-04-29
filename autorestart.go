package autorestart

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tillberg/watcher"
)

func logf(format string, args ...interface{}) {
	log.Printf("[autorestart] "+format+"\n", args...)
}

const errorPath = "*error*"

var _exePath = errorPath

func getExePath() string {
	var err error
	if _exePath == errorPath {
		_exePath, err = exec.LookPath(os.Args[0])
		if err != nil {
			logf("Failed to resolve path to current program: %s", err)
			_exePath = errorPath
		} else {
			_exePath, err = filepath.Abs(_exePath)
			if err != nil {
				logf("Failed to resolve absolute path to current program: %s", err)
				_exePath = errorPath
			} else {
				_exePath = filepath.Clean(_exePath)
			}
		}
	}
	return _exePath
}

// Restart the current program when the program's executable is updated.
// This function is a wrapper around NotifyOnChange and RestartViaExec, calling the
// latter when the former signals that a change was detected.
func RestartOnChange() {
	notifyChan := NotifyOnChange(true)
	<-notifyChan
	logf("%s changed. Restarting via exec.", getExePath())
	// Sort of a maybe-workaround for the issue detailed in RestartViaExec:
	time.Sleep(1 * time.Millisecond)
	RestartViaExec()
}

// Subscribe to a notification when the current process' executable file is modified.
// Returns a channel to which notifications (just `true`) will be sent whenever a
// change is detected.
func NotifyOnChange(usePolling bool) chan bool {
	notifyChan := make(chan bool)
	go func() {
		exePath := getExePath()
		if exePath == errorPath {
			return
		}
		notify, err := watcher.WatchExecutable(exePath, usePolling)
		if err != nil {
			logf("Failed to initialize watcher: %v", err)
			return
		}
		for range notify {
			notifyChan <- true
		}
	}()
	return notifyChan
}

// Restart the current process by calling syscall.Exec, using os.Args (with filepath.LookPath)
// and os.Environ() to recreate the same args & environment that was used when the process was
// originally started.
// Due to using syscall.Exec, this function is not portable to systems that don't support exec.
func RestartViaExec() {
	exePath := getExePath()
	if exePath == errorPath {
		return
	}
	for {
		err := syscall.Exec(exePath, os.Args, os.Environ())
		// Not sure if this is due to user error, a Go regression in 1.5.x, or arch something,
		// but this started failing when called immediately; a short delay (perhaps to switch
		// to a different thread? or maybe to actually delay for some reason?) seems to work
		// all the time. though.
		logf("syscall.Exec failed [%v], trying again in one second...", err)
		time.Sleep(1 * time.Second)
	}
}

func NotifyOnSighup() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	return sigChan
}
