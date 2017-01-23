package utils

import (
	"sync"
	"errors"
)


type Producer interface {
	Produce(Task) error
	Finish()
	SetError(error)
	GetError() error
}

type Consumer interface {
	Consume()
	StopProducer()
	SetError(error)
	GetError() error
}

type Task func(int) error

type ProducerConsumer struct {
	communicationChan     chan Task
	cancel                chan struct{}
	err                   error
	canceled              bool
	cancellationLock      sync.Mutex
	numOfConsumers        int
}

func NewProducerConsumer(numOfConsumers int) *ProducerConsumer {
	consumers := numOfConsumers
	if consumers < 0 {
		consumers = 1
	}
	return &ProducerConsumer{
		communicationChan: make(chan Task),
		cancel: make(chan struct{}),
		numOfConsumers: consumers,
	}
}

func (pc *ProducerConsumer) StopProducer() {
	if !pc.canceled {
		pc.cancellationLock.Lock()
		defer pc.cancellationLock.Unlock()
		if !pc.canceled {
			close(pc.cancel)
			pc.canceled = true
		}
	}
}

func (pc *ProducerConsumer) Finish() {
	close(pc.communicationChan)
}

func (pc *ProducerConsumer) getCommonChannel() <- chan Task {
	return pc.communicationChan
}

func (pc *ProducerConsumer) Produce(t Task) error {
	select {
	case pc.communicationChan <- t:
		return nil
	case <-pc.cancel:
		return errors.New("Producer stopped")
	}
}

func (pc *ProducerConsumer) SetError(e error) {
	pc.err = e
}

func (pc *ProducerConsumer) GetError() error {
	return pc.err
}

func (pc *ProducerConsumer) Consume() {
	var err error
	var wg sync.WaitGroup
	for i := 0; i < pc.numOfConsumers; i++ {
		wg.Add(1)
		go func(threadId int) {
			defer func() {
				wg.Done()
			}()
			for artifactHandler := range pc.getCommonChannel() {
				if err != nil {
					break
				}
				// Must use e and not err due to synchronization issue
				e := artifactHandler(threadId)
				if e != nil {
					err = e
					pc.StopProducer()
					break
				}
			}
		}(i)
	}
	wg.Wait()
}

