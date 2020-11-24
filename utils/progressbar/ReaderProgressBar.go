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

// Initializes a new progress indication for a new file transfer.
// Input: 'total' - file size, 'prefix' - optional description, 'extraInformation' - information to be desplayed.
// Output: A progressBar object

// Used to updated the progress bar progress.
func (p *ReaderProgressBar) ActionWithProgress(args ...interface{}) (results interface{}) {
	if len(args) > 0 {
		reader, ok := args[0].(io.Reader)
		if ok {
			return p.readWithProgress(reader)
		}
	}
	return nil
}

// Abort aborts a progress indication. Called on both successful and unsuccessful operations
func (p *ReaderProgressBar) Abort() {
	// If a replacing bar
	if p.replaceBar != nil {
		// The replacing bar is displayed only if the replacedBar completed, so needs to be dropped only if so
		if p.replaceBar.Completed() {
			p.bar.Abort(true)
		} else {
			p.bar.Abort(false)
		}
	} else {
		close(p.incrChannel)
		p.bar.Abort(true)
	}

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
	if n > 0 && err == nil {
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
