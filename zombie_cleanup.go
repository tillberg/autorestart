package autorestart

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func CleanUpChildZombies() {
	logger := log.New(os.Stderr, "[autorestart.CleanUpZombies] ", log.LstdFlags)
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
					time.Sleep(5 * time.Second)
					if err != nil {
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
				syscall.Wait4(pid, &ws, 0, nil)
				exited = true
			}
		}
	}()
}
