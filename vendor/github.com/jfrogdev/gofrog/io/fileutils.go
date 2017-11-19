package io

import (
	"bufio"
	cr "crypto/rand"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

type RandFile struct {
	*os.File
	Info os.FileInfo
}

const buflen = 4096

var src = rand.NewSource(time.Now().UnixNano())
var rnd = rand.New(src)

func CreateRandomLenFile(maxLen int, filesDir string, prefix string) string {
	file, _ := ioutil.TempFile(filesDir, prefix)
	fname := file.Name()
	len := rnd.Intn(maxLen)
	created, err := CreateRandFile(fname, len)
	if err != nil {
		panic(err)
	}
	defer created.Close()
	//Check that the files were created with expected len
	if created.Info.Size() != int64(len) {
		panic(fmt.Errorf("Unexpected file length. Expected: %d. Got %d.", created.Info.Size(), len))
	}
	return fname
}

func CreateRandFile(path string, len int) (*RandFile, error) {
	f, err := os.Create(path)
	defer f.Close()

	if err != nil {
		return nil, err
	}

	w := bufio.NewWriter(f)
	buf := make([]byte, buflen)

	for i := 0; i <= len; i += buflen {
		cr.Read(buf)
		var wbuflen = buflen
		if i+buflen >= len {
			wbuflen = len - i
		}
		wbuf := buf[0:wbuflen]
		_, err = w.Write(wbuf)
		if err != nil {
			return nil, err
		}
	}
	w.Flush()

	//        if stat, err := file.Stat(); err == nil {

	if info, err := f.Stat(); err != nil {
		return nil, err
	} else {
		file := RandFile{f, info}
		return &file, nil
	}
}
