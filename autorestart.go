package autorestart

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"gopkg.in/fsnotify.v1"
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
	notifyChan := NotifyOnChange()
	<-notifyChan
	logf("%s changed. Restarting via exec.", getExePath())
	// Sort of a maybe-workaround for the issue detailed in RestartViaExec:
	time.Sleep(1 * time.Millisecond)
	RestartViaExec()
}

// Subscribe to a notification when the current process' executable file is modified.
// Returns a channel to which notifications (just `true`) will be sent whenever a
// change is detected.
func NotifyOnChange() chan bool {
	notifyChan := make(chan bool)
	go func() {
		exePath := getExePath()
		if exePath == errorPath {
			return
		}
		exeDir := filepath.Dir(exePath)
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			logf("Failed to initialize gopkg.in/fsnotify.v1 watcher: %s", err)
			return
		}
		abs, _ := filepath.Abs(exeDir)
		err = watcher.Add(abs)
		if err != nil {
			logf("Failed to start filesystem watcher on %s: %s", exeDir, err)
			return
		}
		for {
			select {
			case err := <-watcher.Errors:
				logf("Watcher error: %s", err)
			case ev := <-watcher.Events:
				// log.Println("change", ev.Name, exePath, ev)
				if ev.Name == exePath {
					notifyChan <- true
				}
			}
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
		args := os.Args
		env := os.Environ()
		// logf("calling syscall.Exec with %q, %q, %q", exePath, args, env)
		syscall.Exec(exePath, args, env)
		// Not sure if this is due to user error, a Go regression in 1.5.x, or arch something,
		// but this started failing when called immediately; a short delay (perhaps to switch
		// to a different thread? or maybe to actually delay for some reason?) seems to work
		// all the time. though.
		logf("syscall.Exec failed, trying again in one second...")
		time.Sleep(1 * time.Second)
	}
}
