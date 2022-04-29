//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package autorestart

func CleanUpChildZombies() {
}

func CleanUpChildZombiesQuietly() {
}
