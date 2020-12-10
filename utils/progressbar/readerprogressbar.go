package progressbar

import (
	"io"
	"io/ioutil"
)

type ReaderProgressBar struct {
	*progressBarUnit
	Id int
}

// Wraps an io.Reader for bytes reading tracking
type proxyReader struct {
	unit *progressBarUnit
	io.ReadCloser
}

// Used to update the progress bar progress.
func (p *ReaderProgressBar) ActionWithProgress(reader io.Reader) (results io.Reader) {
	return p.readWithProgress(reader)
}

// Abort aborts a progress indicator. Called on both successful and unsuccessful operations
func (p *ReaderProgressBar) Abort() {
	close(p.incrChannel)
	p.bar.Abort(true)
}

// GetId Returns the ProgressBar ID
func (p *ReaderProgressBar) GetId() (Id int) {
	return p.Id
}

func (p *ReaderProgressBar) getProgressBarUnit() (unit *progressBarUnit) {
	return p.progressBarUnit
}

// Wraps a body of a response (io.Reader) and increments bar accordingly
func (p *ReaderProgressBar) readWithProgress(reader io.Reader) (wrappedReader io.Reader) {
	wrappedReader = initProxyReader(p.progressBarUnit, reader)
	return wrappedReader
}

func initProxyReader(unit *progressBarUnit, reader io.Reader) io.ReadCloser {
	if reader == nil {
		return nil
	}
	rc, ok := reader.(io.ReadCloser)
	if !ok {
		rc = ioutil.NopCloser(reader)
	}
	return &proxyReader{unit, rc}
}

// Overrides the Read method of the original io.Reader.
func (pr *proxyReader) Read(p []byte) (n int, err error) {
	n, err = pr.ReadCloser.Read(p)
	if n > 0 && (err == nil || err == io.EOF) {
		pr.incrChannel(n)
	}
	return
}

func (pr *proxyReader) incrChannel(n int) {
	// When an upload / download error occurs (for example, a bad HTTP error code),
	// The progress bar's Abort method is invoked and closes the channel.
	// Therefore, the channel may be already closed at this stage, which leads to a panic.
	// We therefore need to recover if that happens.
	defer func() {
		recover()
	}()
	pr.unit.incrChannel <- n
}
