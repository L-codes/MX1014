//go:build linux || darwin
// +build linux darwin

package mx1014

import (
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
	if runtime.GOOS == "darwin" && setVal > 24576 {
		setVal = 24576
	}

	rLimit.Max = setVal
	rLimit.Cur = setVal
	unix.Setrlimit(unix.RLIMIT_NOFILE, &rLimit)
}
