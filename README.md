# autorestart

[![GoDoc](https://godoc.org/github.com/tillberg/autorestart?status.png)](http://godoc.org/github.com/tillberg/autorestart)

Go library to restart (via exec) a program when a new version is written to disk.

### Basic usage

To restart daemons in conjunction with [autoinstall][autoinstall], call
`RestartOnChange` during initialization, e.g.:

```go
import "github.com/tillberg/autorestart"
func main() {
    go autorestart.RestartOnChange()
    ...
}
```

This restarts the daemon by calling `syscall.Exec`, which is not portable and is unsupported
on Windows. Exec replaces the current process with the new one, maintaining the same PID
as before and such. This a super-convenient way to restart processes, but it can be a bit
different than what you might expect when being restarted by a parent process; you might want
to read about it a little first: [linux.die.net][linux.die.net] [wikipedia][wikipedia]

Specifically, if your daemon accumulates zombie subprocesses over restarts, the
`CleanUpChildZombies` function might help.

See [documentation][documentation] for additional details.

[documentation]: http://godoc.org/github.com/tillberg/autorestart
[autoinstall]: https://github.com/tillberg/autoinstall
[linux.die.net]: http://linux.die.net/man/3/exec
[wikipedia]: https://en.wikipedia.org/wiki/Exec_(computing)
