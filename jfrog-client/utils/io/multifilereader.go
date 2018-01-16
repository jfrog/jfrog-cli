package io

import (
	"io"
	"os"
	"sort"
)

// Create new multi file ReaderAt
func NewMultiFileReaderAt(filePaths []string) (*multiFileReaderAt, error) {
	readerAt := &multiFileReaderAt{}
	for _, v := range filePaths {
		f, err := os.Open(v)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			return nil, err
		}

		readerAt.filesPaths = append(readerAt.filesPaths, v)
		readerAt.sizeIndex = append(readerAt.sizeIndex, readerAt.size)
		readerAt.size += stat.Size()
	}

	return readerAt, nil
}

type multiFileReaderAt struct {
	filesPaths []string
	size       int64
	sizeIndex  []int64
}

// Get overall size of all the files.
func (multiFileReader *multiFileReaderAt) Size() int64 {
	return multiFileReader.size
}

// ReadAt implementation for multi files
func (multiFileReader *multiFileReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	// Search for the correct index to find the correct file offset
	i := sort.Search(len(multiFileReader.sizeIndex), func(i int) bool { return multiFileReader.sizeIndex[i] > off }) - 1

	readBytes := 0
	for true {
		var f *os.File
		f, err = os.Open(multiFileReader.filesPaths[i])
		if err != nil {
			return
		}
		defer f.Close()
		relativeOff := off + int64(n) - multiFileReader.sizeIndex[i]
		readBytes, err = f.ReadAt(p[n:], relativeOff)
		n += readBytes
		if len(p) == n {
			// Finished reading enough bytes
			return
		}
		if err != nil && err != io.EOF {
			// Error
			return
		}
		if i+1 == len(multiFileReader.filesPaths) {
			// No more files to read from
			return
		}
		// Read from the next file
		i++
	}
	// not suppose to get here
	return
}
