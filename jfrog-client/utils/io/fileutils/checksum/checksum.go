package checksum

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils/checksum/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"hash"
)

type Algorithm int

const (
	MD5    Algorithm = iota
	SHA1
	SHA256
)

var algorithmFunc = map[Algorithm](func() hash.Hash){
	MD5:    md5.New,
	SHA1:   sha1.New,
	SHA256: sha256.New,
}

// Calc all hashes at once using AsyncMultiWriter therefore the file is read only once.
func Calc(reader io.Reader, checksumType ...Algorithm) (map[Algorithm]string, error) {
	hashes := getChecksumByAlgorithm(checksumType ...)
	var multiWriter io.Writer
	pagesize := os.Getpagesize()
	sizedReader := bufio.NewReaderSize(reader, pagesize)
	var hashWriter []io.Writer
	for _, v := range hashes {
		hashWriter = append(hashWriter, v)
	}
	multiWriter = utils.AsyncMultiWriter(hashWriter ...)
	_, err := io.Copy(multiWriter, sizedReader)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	results := sumResults(hashes)
	return results, nil
}

func sumResults(hashes map[Algorithm]hash.Hash) map[Algorithm]string {
	results := map[Algorithm]string{}
	for k, v := range hashes {
		results[k] = fmt.Sprintf("%x", v.Sum(nil))
	}
	return results
}

func getChecksumByAlgorithm(checksumType ...Algorithm) map[Algorithm]hash.Hash {
	hashes := map[Algorithm]hash.Hash{}
	if len(checksumType) == 0 {
		for k, v := range algorithmFunc {
			hashes[k] = v()
		}
		return hashes
	}

	for _, v := range checksumType {
		hashes[v] = algorithmFunc[v]()
	}
	return hashes
}
