package lock

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Lock struct {
	// The current time when the lock was created
	currentTime int64
	// The full path to the lock file.
	fileName string
	pid      int
}

type Locks []Lock

func (locks Locks) Len() int {
	return len(locks)
}

func (locks Locks) Swap(i, j int) {
	locks[i], locks[j] = locks[j], locks[i]
}

func (locks Locks) Less(i, j int) bool {
	return locks[i].currentTime < locks[j].currentTime
}

// Creating a new lock object.
func (lock *Lock) CreateNewLockFile() error {
	lock.currentTime = time.Now().UnixNano()
	folderName, err := lock.createLockDirWithPermissions()
	if err != nil {
		return err
	}
	pid := os.Getpid()
	lock.pid = pid
	err = lock.CreateFile(folderName, pid)
	if err != nil {
		return err
	}
	return nil
}

func (lock *Lock) CreateFile(folderName string, pid int) error {
	// We are creating an empty file with the pid and current time part of the name
	lock.fileName = filepath.Join(folderName, "jfrog-cli.conf.lck."+strconv.Itoa(pid)+"."+strconv.FormatInt(lock.currentTime, 10))
	log.Debug("Creating lock file: ", lock.fileName)
	file, err := os.OpenFile(lock.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return errorutils.CheckError(err)
	}

	if err = file.Close(); err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

func (lock *Lock) createLockDirWithPermissions() (string, error) {
	homeDir, err := config.GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	// The lock created in the lock folder within JFrog CLI Home Dir
	folderName := filepath.Join(homeDir, "lock")
	exists, err := fileutils.IsDirExists(folderName)
	if !exists {
		err = fileutils.CreateDirIfNotExist(folderName)
		if err != nil {
			return "", err
		}

		err = os.Chmod(folderName, 0777)
		if err != nil {
			return "", errorutils.CheckError(err)
		}
	}
	return folderName, nil
}

// Try to acquire a lock
func (lock *Lock) Lock() error {
	filesList, err := lock.getListOfFiles()
	if err != nil {
		return err
	}
	i := 0
	// Trying 1200 times to acquire a lock
	for i <= 1200 {
		// If only one file, means that the process that is running is the one that created the file.
		// We can continue
		if len(filesList) == 1 {
			return nil
		}

		locks, err := lock.getLocks(filesList)
		if err != nil {
			return err
		}
		// If the first timestamp in the sorted locks slice is equal to this timestamp
		// means that the lock can be acquired
		if locks[0].currentTime == lock.currentTime {
			// Edge case, if at the same time (by the nano seconds) two different process created two files.
			// We are checking the PID to know which process can run.
			if locks[0].pid != lock.pid {
				err := lock.removeOtherLockOrWait(locks[0], &filesList)
				if err != nil {
					return err
				}
			} else {
				log.Debug("Lock has been acquired for", lock.fileName)
				return nil
			}
		} else {
			err := lock.removeOtherLockOrWait(locks[0], &filesList)
			if err != nil {
				return err
			}
		}
		i++
	}
	return errors.New("Lock hasn't been acquired.")
}

// Checks if other lock file still exists.
// Or the process that created the lock still running.
func (lock *Lock) removeOtherLockOrWait(otherLock Lock, filesList *[]string) error {
	// Check if file exists.
	exists, err := fileutils.IsFileExists(otherLock.fileName)
	if err != nil {
		return err
	}

	if !exists {
		// Process already finished. Update the list.
		*filesList, err = lock.getListOfFiles()
		if err != nil {
			return err
		}
		return nil
	}
	log.Debug("Lock hasn't been acquired.")

	// Check if the process is running.
	// There are two implementation of the 'isProcessRunning'.
	// One for Windows and one for Unix based systems.
	running, err := isProcessRunning(otherLock.pid)
	if err != nil {
		return err
	}

	if !running {
		log.Debug(fmt.Sprintf("Removing lock file %s since the creating process is no longer running", otherLock.fileName))
		err := otherLock.Unlock()
		// Update list of files
		*filesList, err = lock.getListOfFiles()
		if err != nil {
			return err
		}
		return nil
	}
	// Other process is running. Wait.
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (lock *Lock) getListOfFiles() ([]string, error) {
	// Listing all the files in the lock directory
	filesList, err := fileutils.ListFiles(filepath.Dir(lock.fileName), false)
	if err != nil {
		return nil, err
	}
	return filesList, nil
}

// Returns a list of all available locks.
func (lock *Lock) getLocks(filesList []string) (Locks, error) {
	// Slice of all the timestamps that currently the lock directory has
	var files Locks
	for _, path := range filesList {
		fileName := filepath.Base(path)
		splitted := strings.Split(fileName, ".")

		if len(splitted) != 5 {
			return nil, errorutils.CheckError(fmt.Errorf("Failed while parsing the file name: %s located at: %s. Expecting a different format.", fileName, path))
		}
		// Last element is the timestamp.
		time, err := strconv.ParseInt(splitted[4], 10, 64)
		if err != nil {
			return nil, errorutils.CheckError(err)
		}
		pid, err := strconv.Atoi(splitted[3])
		if err != nil {
			return nil, errorutils.CheckError(err)
		}
		file := Lock{
			currentTime: time,
			pid:         pid,
			fileName:    path,
		}
		files = append(files, file)
	}
	sort.Sort(files)
	return files, nil
}

// Removes the lock file so other process can continue.
func (lock *Lock) Unlock() error {
	log.Debug("Releasing lock: ", lock.fileName)
	exists, err := fileutils.IsFileExists(lock.fileName)
	if err != nil {
		return err
	}

	if exists {
		err = os.Remove(lock.fileName)
		if err != nil {
			return errorutils.CheckError(err)
		}
	}
	return nil
}

func CreateLock() (Lock, error) {
	lockFile := new(Lock)
	err := lockFile.CreateNewLockFile()

	if err != nil {
		return *lockFile, err
	}

	// Trying to acquire a lock for the running process.
	err = lockFile.Lock()
	if err != nil {
		return *lockFile, errorutils.CheckError(err)
	}
	return *lockFile, nil
}
