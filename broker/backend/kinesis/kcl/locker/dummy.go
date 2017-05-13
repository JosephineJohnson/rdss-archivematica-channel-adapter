package locker

import (
	"errors"
	"sync"
	"time"
)

const (
	waitSleep   = time.Duration(100) * time.Millisecond
	waitRetries = 3
)

type DummyReleaser struct {
	locker *DummyLocker
	name   string
}

func NewDummyReleaser(locker *DummyLocker, name string) *DummyReleaser {
	return &DummyReleaser{locker: locker, name: name}
}

func (r *DummyReleaser) Release() error {
	r.locker.mu.Lock()
	defer r.locker.mu.Unlock()
	_, ok := r.locker.ml[r.name]
	if !ok {
		return errors.New("lock could not be found")
	}
	delete(r.locker.ml, r.name)
	return nil
}

// DummyLocker is an implementation of Locker.
type DummyLocker struct {
	ml map[string]bool
	mu sync.Mutex
}

// NewDummyLocker returns a new DummyLocker.
func NewDummyLocker() *DummyLocker {
	return &DummyLocker{
		ml: make(map[string]bool),
	}
}

// LockWait implements Locker.
func (l *DummyLocker) LockWait(name string) (Releaser, error) {
	var err error
	var releaser Releaser
	var success bool

	errTries := 0
	for {
		releaser, success, err = l.Lock(name)
		if err != nil {
			errTries++
			if errTries > waitRetries {
				return nil, err
			}
		} else {
			errTries = 0
		}

		if success == true {
			return releaser, nil
		}

		time.Sleep(waitSleep)
	}
}

// Lock implements Locker.
func (l *DummyLocker) Lock(name string) (Releaser, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.ml[name]
	if ok {
		return nil, false, errors.New("lock already claimed")
	}
	l.ml[name] = true
	return NewDummyReleaser(l, name), true, nil
}
