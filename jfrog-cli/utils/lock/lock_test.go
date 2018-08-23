package lock

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"os"
	"testing"
	"time"
)

func TestLock(t *testing.T) {

	// First creating the first lock object with special pid number that doesn't exists.
	fmt.Print("The process id is: ", os.Getpid())
	b, e := isProcessRunning(os.Getpid()*os.Getpid())
	fmt.Print("Is process exists? ", b, e)
	getLock(os.Getpid()*os.Getpid(), t)
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

	lock := new(Lock)
	err := lock.CreateNewLockFile()
	if err != nil {
		t.Error(err)
	}

	exists, err := fileutils.IsFileExists(lock.fileName)
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Errorf("File %s is missing", lock.fileName)
	}

	lock.Unlock()

	exists, err = fileutils.IsFileExists(lock.fileName)
	if err != nil {
		t.Error(err)
	}

	if exists {
		t.Errorf("File %s exists, but it should have been removed by Unlock", lock.fileName)
	}
}

func TestCreateFile(t *testing.T) {
	pid := os.Getpid()
	lock, folderName := getLock(pid, t)

	exists, err := fileutils.IsFileExists(lock.fileName)
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
		t.Error(fmt.Errorf("Expected one file, got %d. %v", len(files), files))
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
	folderName, err := lock.createLockDirWithPermissions()
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
