package fanout

import (
	"io"
	"github.com/pkg/errors"
)

//A reader that emits its read to multiple consumers using a ReadAll(p []byte) ([]interface{}, error) func
type ReadAllReader struct {
	reader      io.Reader
	consumers   []ReadAllConsumer
	pipeReaders []*io.PipeReader
	pipeWriters []*io.PipeWriter
	results     chan *readerResult
	errs        chan error
}

type ReadAllConsumer interface {
	ReadAll(io.Reader) (interface{}, error)
}

type ReadAllConsumerFunc func(io.Reader) (interface{}, error)

func (f ReadAllConsumerFunc) ReadAll(r io.Reader) (interface{}, error) {
	return f(r)
}

type readerResult struct {
	data interface{}
	pos  int
}

/*
[inr]--r--
          |--w--[pw]--|--[pr]--r
          |--w--[pw]--|--[pr]--r
          |--w--[pw]--|--[pr]--r
*/

func NewReadAllReader(reader io.Reader, consumers ... ReadAllConsumer) *ReadAllReader {
	procLen := len(consumers)
	pipeReaders := make([]*io.PipeReader, procLen)
	pipeWriters := make([]*io.PipeWriter, procLen)
	done := make(chan *readerResult)
	errs := make(chan error)
	//Create pipe r/w for each reader
	for i := 0; i < procLen; i++ {
		pr, pw := io.Pipe()
		pipeReaders[i] = pr
		pipeWriters[i] = pw
	}
	return &ReadAllReader{reader, consumers, pipeReaders, pipeWriters, done, errs}
}

func toWriters(pipeWriters []*io.PipeWriter) (writers []io.Writer) {
	//Convert to an array of io.Writers so it can be taken by a variadic func
	//See: https://groups.google.com/forum/#!topic/golang-nuts/zU3BqD5mKs8
	writers = make([]io.Writer, len(pipeWriters))
	for i, w := range pipeWriters {
		writers[i] = w
	}
	return
}

func (r *ReadAllReader) GetReader(i int) io.Reader {
	return r.pipeReaders[i]
}

func (r *ReadAllReader) ReadAll() ([]interface{}, error) {
	defer close(r.results)
	defer close(r.errs)

	for i, sr := range r.consumers {
		go func(sr ReadAllConsumer, pos int) {
			ret, perr := sr.ReadAll(r.pipeReaders[pos])
			if perr != nil {
				r.errs <- errors.WithStack(perr)
				//panic(perr)
				return
			}
			r.results <- &readerResult{ret, pos}
		}(sr, i)
	}
	go func() {
		defer r.Close()
		mw := io.MultiWriter(toWriters(r.pipeWriters)...)
		_, err := io.Copy(mw, r.reader)
		if err != nil {
			//panic(err)
			r.errs <- errors.WithStack(err)
		}
	}()
	results := make([]interface{}, len(r.consumers))
	for range r.consumers {
		select {
		case err := <-r.errs:
			return nil, err
		case result := <-r.results:
			results[result.pos] = result.data
		}
	}
	return results, nil
}

func (r *ReadAllReader) Close() {
	for _, pw := range r.pipeWriters {
		pw.Close()
	}
}
