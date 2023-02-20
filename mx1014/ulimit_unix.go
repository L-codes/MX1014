//go:build linux || darwin
// +build linux darwin

package mx1014

import (
    "fmt"
    "syscall"
)

func SetUlimit() {
    var rLimit syscall.Rlimit
    err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
    if err != nil {
        fmt.Println("# Error Getting Rlimit ", err)
    }
    if rLimit.Cur < 99999 {
        rLimit.Max = 999999
        rLimit.Cur = 999999

        err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
        if err != nil {
            fmt.Println("# Error Setting Rlimit ", err)
        }
    }
}
