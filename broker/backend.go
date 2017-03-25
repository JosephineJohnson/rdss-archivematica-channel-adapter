package broker

import (
	"context"
	"fmt"
	"strings"
)

// Backend is a low-level interface for communicating with a broker.
type Backend interface {
	Request(context.Context, *Message) error

	RequestResponse(context.Context, *Message) (*Message, error)

	Subscribe(context.Context) error
}

// We have to build a Sender and a Requestor
// See https://tibcoguru.org/tag/jms-queue-requestor/
// See http://www.enterpriseintegrationpatterns.com/patterns/messaging/RequestReplyJmsExample.html

// Constructor is a function that initializes and returns a Broke
// implementation with the given options.
type Constructor func(*Opts) (Backend, error)

var registration = make(map[string]Constructor)

// Opts holds configuration for the broker backend.
// It is meant to be used by implementations of Storage
type Opts struct {
	Opts map[string]string
}

// DialOpts is a daisy-chaining mechanism for setting options to a backend
// during Dial.
type DialOpts func(*Opts) error

// Register register a new broker backend under a name. It is tipically used in
// init functions.
func Register(name string, fn Constructor) error {
	if _, exists := registration[name]; exists {
		return fmt.Errorf("broker backend already exists")
	}
	registration[name] = fn
	return nil
}

// WithOptions parses a string in the format "key1=value1,key2=value2,..." where
// keys and values are specific to each storage backend. Neither key nor value
// may contain the characters "," or "=". Use WithKeyValue repeatedly if these
// characters need to be used.
func WithOptions(options string) DialOpts {
	return func(o *Opts) error {
		pairs := strings.Split(options, ",")
		for _, p := range pairs {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) != 2 {
				return fmt.Errorf("error parsing option %s", kv)
			}
			o.Opts[kv[0]] = kv[1]
		}
		return nil
	}
}

// WithKeyValue sets a key-value pair as option. If called multiple times with
// the same key, the last one wins.
func WithKeyValue(key, value string) DialOpts {
	return func(o *Opts) error {
		o.Opts[key] = value
		return nil
	}
}

// Dial dials the named broker backend using the dial options opts.
func Dial(name string, opts ...DialOpts) (Backend, error) {
	fn, found := registration[name]
	if !found {
		return nil, fmt.Errorf("unknown broker backend type %q", name)
	}
	dOpts := &Opts{Opts: make(map[string]string)}
	var err error
	for _, o := range opts {
		if o != nil {
			err = o(dOpts)
			if err != nil {
				return nil, err
			}
		}
	}
	return fn(dOpts)
}
