package utils

type ErrorsQueue struct {
	errorsChan chan (error)
}

func NewErrorsQueue(size int) *ErrorsQueue {
	queueSize := 1
	if size > 1 {
		queueSize = size
	}
	return &ErrorsQueue{errorsChan: make(chan error, queueSize)}
}

func (errQueue *ErrorsQueue) AddError(err error) {
	select {
	case errQueue.errorsChan <- err:
	default:
		return
	}
}

func (errQueue *ErrorsQueue) GetError() error {
	select {
	case err := <-errQueue.errorsChan:
		return err
	default:
		return nil
	}
}
