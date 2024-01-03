package semaphore

// Semaphore is a type that holds a buffered channel
type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	ch := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		ch <- struct{}{}
	}
	return &Semaphore{ch: ch}
}

func (s *Semaphore) Process(f func() (interface{}, error)) (interface{}, error) {
	select {
	case <-s.ch: // acquire token
		results, err := f()
		s.ch <- struct{}{} // release token
		return results, err
		//default:
		//	return nil, errors.New("semaphore is full")
	}
}

func (s *Semaphore) ProcessLogo(f func() (interface{}, error)) {
	select {
	case <-s.ch: // acquire token
		_, _ = f()
		s.ch <- struct{}{} // release token

		//default:
		//	return nil, errors.New("semaphore is full")
	}
}
