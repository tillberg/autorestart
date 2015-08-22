package autorestart

import (
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
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
			logf("Failed to initialize howeyc/fsnotify watcher: %s", err)
			return
		}
		abs, _ := filepath.Abs(exeDir)
		err = watcher.Watch(abs)
		if err != nil {
			logf("Failed to start filesystem watcher on %s: %s", exeDir, err)
			return
		}
		for {
			select {
			case err := <-watcher.Error:
				logf("Watcher error: %s", err)
			case ev := <-watcher.Event:
				// log.Println("change", ev.Name, exePath, ev)
				if ev.Name == exePath && (ev.IsModify() || ev.IsCreate()) {
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
	syscall.Exec(exePath, os.Args, os.Environ())
}
