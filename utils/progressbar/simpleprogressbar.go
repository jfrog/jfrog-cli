package progressbar

import (
	"io"
)

type SimpleProgressBar struct {
	*progressBarUnit
	Id int
}

// Usesd to update the progress bar progress.
func (p *SimpleProgressBar) ActionWithProgress(reader io.Reader) (results io.Reader) {
	p.bar.Increment()
	return nil
}

// Abort aborts a progress indicator. Called on both successful and unsuccessful operations
func (p *SimpleProgressBar) Abort() {
	p.bar.Abort(true)
}

// GetId Returns the ProgressBar ID
func (p *SimpleProgressBar) GetId() (Id int) {
	return p.Id
}

func (p *SimpleProgressBar) getProgressBarUnit() (unit *progressBarUnit) {
	return p.progressBarUnit
}
