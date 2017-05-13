package checkpointer

import (
	"math/rand"
	"sync"
	"time"
)

const (
	waitSleep   = time.Duration(100) * time.Millisecond
	waitRetries = 3
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// DummyCheckpointer is an implementation of Checkpointer.
type DummyCheckpointer struct {
	keys map[string]string
	mu   sync.Mutex
}

// NewDummyCheckpointer returns a new DummyCheckpointer.
func NewDummyCheckpointer() *DummyCheckpointer {
	return &DummyCheckpointer{keys: make(map[string]string)}
}

// GetCheckpoint implements DummyCheckpointer.
func (c *DummyCheckpointer) GetCheckpoint(key string) (string, error) {
	errTries := 0
	for {
		val, err := c.get(key)
		if err != nil {
			errTries++
			if errTries > waitRetries {
				return "", err
			}
			time.Sleep(waitSleep)
			continue
		}
		return val, nil
	}
}

// SetCheckpoint implements DummyCheckpointer.
func (c *DummyCheckpointer) SetCheckpoint(key, value string) error {
	errTries := 0
	for {
		err := c.set(key, value)
		if err != nil {
			errTries++
			if errTries > waitRetries {
				return err
			}
			time.Sleep(waitSleep)
			continue
		}
		return nil
	}
}

func (c *DummyCheckpointer) get(key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.keys[key]
	if ok {
		return v, nil
	}
	// It should be a non-empty string if I already had a checkpoint stored. But
	// this implementation is not persistent. The convention is to return an
	// empty string which will cause a TRIM_HORIZON iterator to be used to start
	// reading from the last untrimmed record.
	c.keys[key] = ""
	return "", nil
}

func (c *DummyCheckpointer) set(key, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys[key] = value
	return nil
}
