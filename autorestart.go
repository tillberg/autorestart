package autorestart

import (
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const logPrefix = "@(dim:[autorebuild]) "

func RestartOnChange() {
	logger := log.New(os.Stderr, "[autorestart.RestartOnChange] ", log.LstdFlags)
	exePath, err := exec.LookPath(os.Args[0])
	if err != nil {
		logger.Printf("Failed to resolve path to current program: %s\n", err)
		return
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		logger.Printf("Failed to resolve absolute path to current program: %s\n", err)
	}
	exePath = filepath.Clean(exePath)
	exeDir := filepath.Dir(exePath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Printf("Failed to initialize howeyc/fsnotify watcher: %s\n", err)
		return
	}
	abs, _ := filepath.Abs(exeDir)
	err = watcher.Watch(abs)
	if err != nil {
		logger.Printf("Failed to start filesystem watcher on %s: %s\n", exeDir, err)
	}
	for {
		select {
		case err := <-watcher.Error:
			logger.Printf("Watcher error: %s\n", err)
		case ev := <-watcher.Event:
			// log.Println("change", ev.Name, exePath, ev)
			if ev.Name == exePath && (ev.IsModify() || ev.IsCreate()) {
				logger.Printf("%s changed. Restarting via exec.\n", exePath)
				syscall.Exec(exePath, os.Args, os.Environ())
			}
		}
	}
}
