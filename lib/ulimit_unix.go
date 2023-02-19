//go:build darwin || linux

package lib

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
