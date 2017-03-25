package broker_test

import (
	"context"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
)

type exampleRequester struct {
	backend broker.Backend
}

func (c *exampleRequester) SendDocument() {
	c.backend.Request(context.TODO(), &broker.Message{})
}

// ExampleRequester illustrates how to implement a requester using an
// implementation of a broker.Backend.
func ExampleRequester() {
	b, err := broker.Dial("kinesis")
	if err == nil {
		panic("pppuuuuuuumm!")
	}
	c := &exampleRequester{backend: b}
	c.SendDocument()
}
