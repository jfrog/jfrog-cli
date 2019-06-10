package lock

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"math"
	"os"
	"testing"
	"time"
)

/*
	The lock mechanism prefers earlier lock requests. If two locks requests have same time stamps, it'll take the one with the smaller PID first.
	Here we test the functionality of a real process with a real PID and a dummy process with MaxInt pid.
*/
func TestLockSmallerPid(t *testing.T) {
	log.SetDefaultLogger()
	// First creating the first lock object with special pid number that doesn't exists.
	firstLock, _ := getLock(math.MaxInt32, t)
	// Creating a second lock object with the running PID
	secondLock, folderName := getLock(os.Getpid(), t)

	// Confirming that only two locks are located in the lock directory
	files, err := fileutils.ListFiles(folderName, false)
	if err != nil {
		t.Error(err)
	}

	if len(files) != 2 {
		t.Error("Expected 2 files but got ", len(files))
	}

	// Performing lock. This should work since the first lock PID is not running. The Lock() will remove it.
	err = secondLock.Lock()
	if err != nil {
		t.Error(err)
	}
	// Unlocking to remove the lock file.
	err = secondLock.Unlock()
	if err != nil {
		t.Error(err)
	}

	// If timestamp equals, secondLock.Lock() is not expected to delete first lock's file, since os.Getpid() < math.MaxInt32.
	if firstLock.currentTime == secondLock.currentTime {
		err = firstLock.Unlock()
		if err != nil {
			t.Error(err)
		}
	}

	// Confirming that no locks are located in the lock directory
	files, err = fileutils.ListFiles(folderName, false)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 0 {
		t.Error("Expected 0 files but got", len(files), files)
	}
}

/*
	The lock mechanism prefers earlier lock requests. If two locks requests have same time stamps, it'll take the one with the smaller PID first.
	Here we test the functionality of a real process with a real PID and a dummy process with -1 pid.
*/
func TestLockBiggerPid(t *testing.T) {
	log.SetDefaultLogger()
	// First creating the first lock object with special pid number that doesn't exists.
	getLock(-1, t)
	// Creating a second lock object with the running PID
	secondLock, folderName := getLock(os.Getpid(), t)

	// Confirming that only two locks are located in the lock directory
	files, err := fileutils.ListFiles(folderName, false)
	if err != nil {
		t.Error(err)
	}

	if len(files) != 2 {
		t.Error("Expected 2 files but got ", len(files), files)
	}

	// Performing lock. This should work since the first lock PID is not running. The Lock() will remove it.
	err = secondLock.Lock()
	if err != nil {
		t.Error(err)
	}
	// Unlocking to remove the lock file.
	err = secondLock.Unlock()
	if err != nil {
		t.Error(err)
	}

	// Confirming that no locks are located in the lock directory
	files, err = fileutils.ListFiles(folderName, false)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 0 {
		t.Error("Expected 0 files but got", len(files), files)
	}
}

func TestUnlock(t *testing.T) {
	log.SetDefaultLogger()
	lock := new(Lock)
	err := lock.CreateNewLockFile()
	if err != nil {
		t.Error(err)
	}

	exists, err := fileutils.IsFileExists(lock.fileName, false)
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Errorf("File %s is missing", lock.fileName)
	}

	lock.Unlock()

	exists, err = fileutils.IsFileExists(lock.fileName, false)
	if err != nil {
		t.Error(err)
	}

	if exists {
		t.Errorf("File %s exists, but it should have been removed by Unlock", lock.fileName)
	}
}

func TestCreateFile(t *testing.T) {
	log.SetDefaultLogger()
	pid := os.Getpid()
	lock, folderName := getLock(pid, t)

	exists, err := fileutils.IsFileExists(lock.fileName, false)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !exists {
		t.Error("Lock wan't created.")
		t.FailNow()
	}

	files, err := fileutils.ListFiles(folderName, false)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(files) != 1 {
		t.Error(fmt.Errorf("Expected one file, got %d.", len(files)))
		t.FailNow()
	}

	if files[0] != lock.fileName {
		t.Error(fmt.Errorf("Expected filename %s, got %s", lock.fileName, files[0]))
	}
	// Removing the created lock file
	err = lock.Unlock()
	if err != nil {
		t.Error(err)
	}
}

func getLock(pid int, t *testing.T) (Lock, string) {
	currentTime := time.Now().UnixNano()
	lock := Lock{
		pid:         pid,
		currentTime: currentTime,
	}
	folderName, err := CreateLockDir()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = lock.CreateFile(folderName, pid)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	return lock, folderName
}
