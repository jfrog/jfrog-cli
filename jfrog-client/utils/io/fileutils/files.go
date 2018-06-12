package fileutils

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils/checksum"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const SYMLINK_FILE_CONTENT = ""

var tempDirPath string

func GetFileSeparator() string {
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
	if err != nil {
		return false, errorutils.CheckError(err)
	}
	return !f.IsDir(), nil
}

func IsPathSymlink(path string) bool {
	f, _ := os.Lstat(path)
	return f != nil && IsFileSymlink(f)
}

func IsFileSymlink(file os.FileInfo) bool {
	return file.Mode()&os.ModeSymlink != 0
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
		fileName = path[index+1:]
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
		localTargetPath = filepath.Join(targetDirPath, relativePath)
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
	sep := GetFileSeparator()
	if !strings.HasSuffix(path, sep) {
		path += sep
	}
	fileList := []string{}
	files, _ := ioutil.ReadDir(path)
	path = strings.TrimPrefix(path, "."+sep)

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
	if file == nil {
		return bytes.NewBuffer([]byte(SYMLINK_FILE_CONTENT))
	}
	return bufio.NewReader(file)
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
		fileName = filepath.Join(localPath, fileName)
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
	defer file.Close()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	fileInfo, err := file.Stat()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	details.Size = fileInfo.Size()
	return details, nil
}

func calcChecksumDetails(filePath string) (ChecksumDetails, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if errorutils.CheckError(err) != nil {
		return ChecksumDetails{}, err
	}
	checksumInfo, err := checksum.Calc(file)
	if err != nil {
		return ChecksumDetails{}, err
	}
	return ChecksumDetails{Md5: checksumInfo[checksum.MD5], Sha1: checksumInfo[checksum.SHA1], Sha256: checksumInfo[checksum.SHA256]}, nil
}

type FileDetails struct {
	Checksum ChecksumDetails
	Size     int64
}

type ChecksumDetails struct {
	Md5    string
	Sha1   string
	Sha256 string
}

func CopyFile(dst, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	fileName, _ := GetFileAndDirFromPath(src)
	dstPath, err := CreateFilePath(dst, fileName)
	if err != nil {
		return err
	}
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	io.Copy(dstFile, srcFile)
	return nil
}

// Copy directory content from one path to another.
// includeDirs means to copy also the dirs if presented in the src folder.
func CopyDir(fromPath, toPath string, includeDirs bool) error {
	err := CreateDirIfNotExist(toPath)
	if err != nil {
		return err
	}

	files, err := ListFiles(fromPath, includeDirs)
	if err != nil {
		return err
	}

	for _, v := range files {
		dir, err := IsDir(v)
		if err != nil {
			return err
		}

		if dir {
			toPath := toPath + GetFileSeparator() + filepath.Base(v)
			err := CopyDir(v, toPath, true)
			if err != nil {
				return err
			}
			continue
		}
		err = CopyFile(toPath, v)
		if err != nil {
			return err
		}
	}
	return err
}