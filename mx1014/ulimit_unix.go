//go:build linux || darwin
// +build linux darwin

package mx1014

import (
	"log"
	"runtime"

	"golang.org/x/sys/unix"
)

func SetUlimit() {
	const max = uint64(999999)

	var rLimit unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &rLimit); err != nil {
		return
	}
	if rLimit.Cur >= max {
		return
	}

	setVal := max
	if runtime.GOOS == "darwin" && rLimit.Cur > 24576 {
		return
	}

	old := rLimit.Cur
	rLimit.Max = setVal
	rLimit.Cur = setVal
	if err := unix.Setrlimit(unix.RLIMIT_NOFILE, &rLimit); err == nil {
		log.Printf("# Set ulimit %d to %d\n", old, setVal)
	}
}
