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

See [documentation][documentation] for additional details.

[documentation]: http://godoc.org/github.com/tillberg/autorestart
[autoinstall]: https://github.com/tillberg/autoinstall
