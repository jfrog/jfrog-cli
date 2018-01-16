package io

import (
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func TestNewMultiFileReaderAt(t *testing.T) {
	tests := []struct {
		name         string
		filesContent [][]byte
		offset       int64
		expected     []byte
	}{
		{name: "simple1", filesContent: [][]byte{{1, 2, 3}}, offset: 0, expected: []byte{1, 2}},
		{name: "simple2", filesContent: [][]byte{{1, 2, 3}}, offset: 4, expected: []byte{}},
		{name: "simple3", filesContent: [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, offset: 4, expected: []byte{5, 6, 7, 8}},
		{name: "front", filesContent: [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, offset: 0, expected: []byte{1, 2, 3, 4}},
		{name: "all", filesContent: [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, offset: 0, expected: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{name: "back1", filesContent: [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, offset: 8, expected: []byte{9}},
		{name: "back2", filesContent: [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, offset: 9, expected: []byte{}},
		{name: "back3", filesContent: [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, offset: 8, expected: []byte{9, 0, 0}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files := []string{}
			// Create file with content
			for k, v := range test.filesContent {
				f, err := ioutil.TempFile("", strconv.Itoa(k))
				if err != nil {
					t.Error(err)
				}
				_, err = f.Write(v)
				if err != nil {
					t.Error(err)
				}
				files = append(files, f.Name())
				f.Close()
				defer os.Remove(f.Name())
			}

			// Create multiFileReaderAt
			multiReader, err := NewMultiFileReaderAt(files)
			if err != nil {
				t.Error(err)
			}

			buf := make([]byte, len(test.expected))
			n, err := multiReader.ReadAt(buf, test.offset)

			// Validate results
			if err != nil && err != io.EOF {
				t.Error(err)
			}
			if n != len(test.expected) && err != io.EOF {
				t.Error("Expected n:", len(test.expected), "got:", n)
			}
			if !reflect.DeepEqual(test.expected, buf) {
				t.Error("Expected:", test.expected, "got:", buf)
			}
		})
	}
}
