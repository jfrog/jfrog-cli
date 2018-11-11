// +build linux darwin freebsd

package lock

import (
	"syscall"
	"os"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// This file will be compiled only on unix systems.
// Checks if the process is running.
// If error occurs, check if the error is part of the OS permission errors. This means the process is running.
// Else means the process is not running.
func isProcessRunning(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	err = process.Signal(syscall.Signal(0))
	// If err is not nil, then the other process might still be running.
	if err != nil {
		// If this is a permission error, then the other process is running
		if os.IsPermission(err) {
			log.Debug("Other process still alive: ", err.Error())
			return true, nil
		}
		// The other process died without unlocking. Let's unlock.
		return false, nil
	}
	// This means there were no errors, same process.
	return true, nil
}
