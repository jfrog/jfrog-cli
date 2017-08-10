package fileutils

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"runtime"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/types"
	"path"
	"sync"
	"path/filepath"
	"archive/zip"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
)

const SYMLINK_FILE_CONTENT = ""

var tempDirPath string

func GetFileSeperator() string {
	if runtime.GOOS == "windows" {
		return "\\"
	}
	return "/"
}

func IsPathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsFileExists(path string) (bool, error) {
	if !IsPathExists(path) {
		return false, nil
	}
	f, err := os.Stat(path)
	err = errorutils.CheckError(err)
	if err != nil {
		return false, err
	}
	return !f.IsDir(), nil
}

func IsPathSymlink(path string) bool {
	f, _ := os.Lstat(path)
	return f != nil && IsFileSymlink(f)
}

func IsFileSymlink(file os.FileInfo) bool {
	return file.Mode() & os.ModeSymlink != 0
}

func IsDir(path string) (bool, error) {
	if !IsPathExists(path) {
		return false, nil
	}
	f, err := os.Stat(path)
	err = errorutils.CheckError(err)
	if err != nil {
		return false, err
	}
	return f.IsDir(), nil
}

func GetFileAndDirFromPath(path string) (fileName, dir string) {
	index1 := strings.LastIndex(path, "/")
	index2 := strings.LastIndex(path, "\\")
	var index int
	if index1 >= index2 {
		index = index1
	} else {
		index = index2
	}
	if index != -1 {
		fileName = path[index + 1:]
		dir = path[:index]
		return
	}
	fileName = path
	dir = ""
	return
}

// Get the local path and filename from original file name and path according to targetPath
func GetLocalPathAndFile(originalFileName, relativePath, targetPath string, flat bool) (localTargetPath, fileName string) {
	targetFileName, targetDirPath := GetFileAndDirFromPath(targetPath)
	localTargetPath = targetDirPath
	if !flat {
		localTargetPath = path.Join(targetDirPath, relativePath)
	}

	fileName = originalFileName
	if targetFileName != "" {
		fileName = targetFileName
	}
	return
}

// Return the recursive list of files and directories in the specified path
func ListFilesRecursiveWalkIntoDirSymlink(path string, walkIntoDirSymlink bool) (fileList []string, err error) {
	fileList = []string{}
	err = Walk(path, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	}, walkIntoDirSymlink)
	err = errorutils.CheckError(err)
	return
}

// Return the list of files and directories in the specified path
func ListFiles(path string, includeDirs bool) ([]string, error) {
	sep := GetFileSeperator()
	if !strings.HasSuffix(path, sep) {
		path += sep
	}
	fileList := []string{}
	files, _ := ioutil.ReadDir(path)
	path = strings.TrimPrefix(path, "." + sep)

	for _, f := range files {
		filePath := path + f.Name()
		exists, err := IsFileExists(filePath)
		if err != nil {
			return nil, err
		}
		if exists || IsPathSymlink(filePath) {
			fileList = append(fileList, filePath)
		} else if includeDirs {
			isDir, err := IsDir(filePath)
			if err != nil {
				return nil, err
			}
			if isDir {
				fileList = append(fileList, filePath)
			}
		}
	}
	return fileList, nil
}

func GetUploadRequestContent(file *os.File) io.Reader {
	var reqBody io.Reader
	reqBody = file
	if file == nil {
		reqBody = bytes.NewBuffer([]byte(SYMLINK_FILE_CONTENT))
	}
	return reqBody
}

func GetFileSize(file *os.File) (int64, error) {
	size := int64(0)
	if file != nil {
		fileInfo, err := file.Stat()
		if errorutils.CheckError(err) != nil {
			return size, err
		}
		size = fileInfo.Size()
	}
	return size, nil
}

func CreateFilePath(localPath, fileName string) (string, error) {
	if localPath != "" {
		err := os.MkdirAll(localPath, 0777)
		if errorutils.CheckError(err) != nil {
			return "", err
		}
		fileName = localPath + "/" + fileName
	}
	return fileName, nil
}

func CreateDirIfNotExist(path string) error {
	exist, err := IsDir(path)
	if exist || err != nil {
		return err
	}
	_, err = CreateFilePath(path, "")
	return err
}

func GetTempDirPath() (string, error) {
	if tempDirPath == "" {
		err := errorutils.CheckError(errors.New("Function cannot be used before 'tempDirPath' is created."))
		if err != nil {
			return "", err
		}
	}
	return tempDirPath, nil
}

func CreateTempDirPath() error {
	if tempDirPath != "" {
		err := errorutils.CheckError(errors.New("'tempDirPath' has already been initialized."))
		if err != nil {
			return err
		}
	}
	path, err := ioutil.TempDir("", "jfrog.cli.")
	err = errorutils.CheckError(err)
	if err != nil {
		return err
	}
	tempDirPath = path
	return nil
}

func RemoveTempDir() error {
	defer func() {
		tempDirPath = ""
	}()

	exists, err := IsDirExists(tempDirPath)
	if err != nil {
		return err
	}
	if exists {
		return os.RemoveAll(tempDirPath)
	}
	return nil
}

func IsDirExists(path string) (bool, error) {
	if !IsPathExists(path) {
		return false, nil
	}
	f, err := os.Stat(path)
	err = errorutils.CheckError(err)
	if err != nil {
		return false, err
	}
	return f.IsDir(), nil
}

// Reads the content of the file in the source path and appends it to
// the file in the destination path.
func AppendFile(srcPath string, destFile *os.File) error {
	srcFile, err := os.Open(srcPath)
	err = errorutils.CheckError(err)
	if err != nil {
		return err
	}

	defer func() error {
		err := srcFile.Close()
		return errorutils.CheckError(err)
	}()

	reader := bufio.NewReader(srcFile)

	writer := bufio.NewWriter(destFile)
	buf := make([]byte, 1024000)
	for {
		n, err := reader.Read(buf)
		if err != io.EOF {
			err = errorutils.CheckError(err)
			if err != nil {
				return err
			}
		}
		if n == 0 {
			break
		}
		_, err = writer.Write(buf[:n])
		err = errorutils.CheckError(err)
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	return errorutils.CheckError(err)
}

func GetHomeDir() string {
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}
	user, err := user.Current()
	if err == nil {
		return user.HomeDir
	}
	return ""
}

func ReadFile(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	err = errorutils.CheckError(err)
	return content, err
}

func GetFileDetails(filePath string) (*FileDetails, error) {
	var err error
	details := new(FileDetails)
	details.Checksum, err = calcChecksumDetails(filePath)

	file, err := os.Open(filePath)
	err = errorutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	err = errorutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	details.Size = fileInfo.Size()
	return details, nil
}

func calcChecksumDetails(filePath string) (ChecksumDetails, error) {
	checksumArray := []struct {
		calc  func(string) (string, error)
		value string
		err   error
	}{{calc: CalcMd5}, {calc: CalcSha1}}

	var wg sync.WaitGroup
	for i, checksum := range checksumArray {
		wg.Add(1)
		go func(i int, calc func(string) (string, error)) {
			checksumArray[i].value, checksumArray[i].err = calc(filePath)
			wg.Done()
		}(i, checksum.calc)
	}
	wg.Wait()

	for _, checksum := range checksumArray {
		if checksum.err != nil {
			return ChecksumDetails{}, checksum.err
		}
	}
	return ChecksumDetails{Md5: checksumArray[0].value, Sha1: checksumArray[1].value}, nil
}

func CalcSha1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	errorutils.CheckError(err)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return GetSha1(file)
}

func GetSha1(input io.Reader) (string, error) {
	var resSha1 []byte
	hashSha1 := sha1.New()
	_, err := io.Copy(hashSha1, input)
	err = errorutils.CheckError(err)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hashSha1.Sum(resSha1)), nil
}

func CalcMd5(filePath string) (string, error) {
	var err error
	file, err := os.Open(filePath)
	err = errorutils.CheckError(err)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return GetMd5(file)
}

func GetMd5(input io.Reader) (string, error) {
	var resMd5 []byte
	hashMd5 := md5.New()
	_, err := io.Copy(hashMd5, input)
	err = errorutils.CheckError(err)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hashMd5.Sum(resMd5)), nil
}

func ZipFolderFiles(source, target string) (err error) {
	zipFile, err := os.Create(target)
	if err != nil {
		errorutils.CheckError(err)
		return
	}
	defer func() {
		if cerr := zipFile.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	archive := zip.NewWriter(zipFile)
	defer func() {
		if cerr := archive.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) (currentErr error) {
		if info.IsDir() {
			return
		}

		if err != nil {
			currentErr = err
			return
		}

		header, currentErr := zip.FileInfoHeader(info)
		if currentErr != nil {
			errorutils.CheckError(currentErr)
			return
		}

		header.Method = zip.Deflate
		writer, currentErr := archive.CreateHeader(header)
		if currentErr != nil {
			errorutils.CheckError(currentErr)
			return
		}

		file, currentErr := os.Open(path)
		if currentErr != nil {
			errorutils.CheckError(currentErr)
			return
		}
		defer func() {
			if cerr := file.Close(); cerr != nil && currentErr == nil {
				currentErr = cerr
			}
		}()
		_, currentErr = io.Copy(writer, file)
		return
	})
	return
}

type FileDetails struct {
	Checksum     ChecksumDetails
	Size         int64
	AcceptRanges *types.BoolEnum
}

type ChecksumDetails struct {
	Md5          string
	Sha1         string
}