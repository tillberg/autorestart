//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package autorestart

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func cleanUpChildZombiesInt(quiet bool) {
	var logWriter io.Writer
	if quiet {
		logWriter = ioutil.Discard
	} else {
		logWriter = os.Stderr
	}
	logger := log.New(logWriter, "[autorestart.CleanUpZombies] ", log.LstdFlags)
	cmd := exec.Command("pgrep", "-P", fmt.Sprintf("%d", os.Getpid()))
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	err := cmd.Start()
	if err != nil {
		logger.Println("Error starting pgrep in CleanUpZombieChildren", err)
	}
	err = cmd.Wait()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			logger.Println("Failed to execute pgrep in CleanUpZombieChildren", err)
		}
	}
	go func() {
		scanner := bufio.NewScanner(&outBuf)
		for scanner.Scan() {
			line := scanner.Text()
			pid64, err := strconv.ParseInt(line, 10, 32)
			if err != nil {
				logger.Println("Could not parse PID from line", line, "in CleanUpZombieChildren", err)
				continue
			}
			pid := int(pid64)
			if pid != cmd.Process.Pid {
				logger.Printf("Cleaning up zombie child with PID %d.\n", pid)
				exited := false
				go func() {
					// Send SIGHUP, SIGINT, and finally SIGTERM, on long delays, to encourage still-living
					// child processes to draw closer to the netherworld.
					time.Sleep(5 * time.Second)
					if exited {
						return
					}
					logger.Printf("Sending SIGHUP to %d.\n", pid)
					err := syscall.Kill(pid, syscall.SIGHUP)
					if err != nil {
						return
					}
					time.Sleep(120 * time.Second)
					if exited {
						return
					}
					logger.Printf("Sending SIGINT to %d.\n", pid)
					err = syscall.Kill(pid, syscall.SIGINT)
					if err != nil {
						return
					}
					time.Sleep(60 * time.Second)
					if exited {
						return
					}
					logger.Printf("Sending SIGTERM to %d.\n", pid)
					syscall.Kill(pid, syscall.SIGTERM)
				}()
				ws := syscall.WaitStatus(0)
				_, err := syscall.Wait4(pid, &ws, 0, nil)
				if err != nil {
					logger.Printf("Error while waiting for PID %d to exit: %s\n", pid, err)
				} else {
					logger.Printf("Zombie %d has gone to rest.\n", pid)
				}
				exited = true
			}
		}
	}()
}

// This is a utility function to clean up zombie subprocesses that can get left behind by
// RestartOnChange, due to the fashion in which it restarts the process. This will synchronously
// execute `pgrep` to get a list of child process IDs, and then asychronously call syscall.Wait4
// on each one to remove zombies. In addition, if those processes are not yet terminated, this
// calls SIGHUP, SIGINT, and finally SIGTERM over the course of a few minutes in order to
// encourage them to die.
func CleanUpChildZombies() {
	cleanUpChildZombiesInt(false)
}

// This is identical to CleanUpChildZombies except that it outputs nothing to stderr.
func CleanUpChildZombiesQuietly() {
	cleanUpChildZombiesInt(true)
}
