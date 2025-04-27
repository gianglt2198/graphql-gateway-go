package async

import (
	"sync"
	"sync/atomic"
)

// WaitAll waits for all the given errables to finish, and returns
// the last error occurred in all errables, if any.
func WaitAll(chans ...<-chan error) error {
	var wg sync.WaitGroup
	wg.Add(len(chans))

	var lastErr atomic.Value
	for _, ch := range chans {
		go func(ch <-chan error) {
			defer wg.Done()
			if err, open := <-ch; open {
				if err != nil {
					lastErr.Store(err)
				}
			} else {
				return
			}
		}(ch)
	}

	wg.Wait()

	if lastErr.Load() == nil {
		return nil
	}
	return lastErr.Load().(error)
}

func Errable(fn func() error) <-chan error {
	ch := make(chan error)
	go func() {
		ch <- fn()
		close(ch)
	}()
	return ch
}
