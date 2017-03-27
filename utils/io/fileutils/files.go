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
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/types"
	"path"
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
	err = cliutils.CheckError(err)
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
	err = cliutils.CheckError(err)
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
	err = cliutils.CheckError(err)
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
		if cliutils.CheckError(err) != nil {
			return size, err
		}
		size = fileInfo.Size()
	}
	return size, nil
}

func CreateFilePath(localPath, fileName string) (string, error) {
	if localPath != "" {
		err := os.MkdirAll(localPath, 0777)
		if cliutils.CheckError(err) != nil {
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
	return  err
}

func GetTempDirPath() (string, error) {
	if tempDirPath == "" {
		err := cliutils.CheckError(errors.New("Function cannot be used before 'tempDirPath' is created."))
		if err != nil {
			return "", err
		}
	}
	return tempDirPath, nil
}

func CreateTempDirPath() error {
	if tempDirPath != "" {
		err := cliutils.CheckError(errors.New("'tempDirPath' has already been initialized."))
		if err != nil {
			return err
		}
	}
	path, err := ioutil.TempDir("", "jfrog.cli.")
	err = cliutils.CheckError(err)
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
	err = cliutils.CheckError(err)
	if err != nil {
		return false, err
	}
	return f.IsDir(), nil
}

// Reads the content of the file in the source path and appends it to
// the file in the destination path.
func AppendFile(srcPath string, destFile *os.File) error {
	srcFile, err := os.Open(srcPath)
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}

	defer func() error {
		err := srcFile.Close()
		return cliutils.CheckError(err)
	}()

	reader := bufio.NewReader(srcFile)

	writer := bufio.NewWriter(destFile)
	buf := make([]byte, 1024000)
	for {
		n, err := reader.Read(buf)
		if err != io.EOF {
			err = cliutils.CheckError(err)
			if err != nil {
				return err
			}
		}
		if n == 0 {
			break
		}
		_, err = writer.Write(buf[:n])
		err = cliutils.CheckError(err)
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	return cliutils.CheckError(err)
}

func GetHomeDir() string {
	user, err := user.Current()
	if err == nil {
		return user.HomeDir
	}
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}
	return ""
}

func ReadFile(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	err = cliutils.CheckError(err)
	return content, err
}

func GetFileDetails(filePath string) (*FileDetails, error) {
	var err error
	details := new(FileDetails)
	details.Md5, err = CalcMd5(filePath)
	if err != nil {
		return nil, err
	}
	details.Sha1, err = CalcSha1(filePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	err = cliutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	err = cliutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	details.Size = fileInfo.Size()
	return details, nil
}

func CalcSha1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	cliutils.CheckError(err)
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
	err = cliutils.CheckError(err)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hashSha1.Sum(resSha1)), nil
}

func CalcMd5(filePath string) (string, error) {
	var err error
	file, err := os.Open(filePath)
	err = cliutils.CheckError(err)
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
	err = cliutils.CheckError(err)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hashMd5.Sum(resMd5)), nil
}

type FileDetails struct {
	Md5          string
	Sha1         string
	Size         int64
	AcceptRanges *types.BoolEnum
}


