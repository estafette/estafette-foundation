package foundation

type Semaphore interface {
	// Acquire tries to get a lock and blocks until it does
	Acquire()
	// Release releases a lock
	Release()
	// Wait until all locks are released
	Wait()
}

type semaphore struct {
	semaphoreChannel chan struct{}
}

func NewSemaphore(maxConcurrency int) Semaphore {
	return &semaphore{
		semaphoreChannel: make(chan struct{}, maxConcurrency),
	}
}

func (s *semaphore) Acquire() {
	s.semaphoreChannel <- struct{}{}
}

func (s *semaphore) Release() {
	<-s.semaphoreChannel
}

func (s *semaphore) Wait() {
	for i := 0; i < cap(s.semaphoreChannel); i++ {
		s.Acquire()
	}

	// reset so the semaphore can be used again
	s.semaphoreChannel = make(chan struct{}, cap(s.semaphoreChannel))
}
