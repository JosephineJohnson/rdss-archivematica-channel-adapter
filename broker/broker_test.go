package broker_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

func Example() {
	// Create context
	ctx := context.Background()

	// Set up a new backend
	var opts []backend.DialOpts
	backend, err := backend.Dial("backendmock", opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer backend.Close()
	c, err := broker.New(backend, nil, &broker.Config{QueueError: "error", QueueInvalid: "invalid", QueueMain: "main"})
	if err != nil {
		log.Fatal(err)
	}

	// Send a CreateMetadata request with a timeout of five seconds.
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	err = c.Metadata.Create(ctx, &message.MetadataCreateRequest{})
	if err != nil {
		log.Fatal(err)
	}

	// A publisher can publish a MetadataCreate request.
	err = c.Metadata.Create(ctx, &message.MetadataCreateRequest{})
	if err != nil {
		panic(err)
	}

	// A subscriber can subscribe to MetadataCreate requests. If you return an
	// error from the handler the message will... TODO!
	c.SubscribeType(message.TypeMetadataCreate, func(msg *message.Message) error {
		fmt.Println("MetadataCreate received!")
		return nil
	})
}
