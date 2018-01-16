package broker

import (
	"errors"

	"github.com/spf13/cast"
)

type Config struct {
	QueueMain        string
	QueueInvalid     string
	QueueError       string
	RepositoryConfig *RepositoryConfig
	Validation       ValidationMode
}

type RepositoryConfig struct {
	Backend          string
	DynamoDBTLS      bool
	DynamoDBRegion   string
	DynamoDBEndpoint string
	DynamoDBTable    string
}

func (c *Config) Validate() error {
	if c.QueueError == "" {
		return errors.New("error queue name is undefined")
	}
	if c.QueueInvalid == "" {
		return errors.New("invalid queue name is undefined")
	}
	if c.QueueMain == "" {
		return errors.New("main queue name is undefined")
	}
	if c.RepositoryConfig == nil || c.RepositoryConfig.Backend == "" {
		return errors.New("repository config is undefined")
	}
	if c.RepositoryConfig.Backend == "dynamodb" && c.RepositoryConfig.DynamoDBTable == "" {
		return errors.New("repository config is missing details")
	}
	return nil
}

// ValidationMode determines the type of message validation that the client
// is going to perform when new messages are received.
type ValidationMode int

const (
	// Messages are rejected if invalid, validation issues will be logged.
	ValidationModeStrict ValidationMode = iota

	// Messages will not be rejected but the validation issues will be logged.
	ValidationModeWarnings

	// Message validator is disabled.
	ValidationModeDisabled
)

func (c *Config) SetValidationMode(mode string) {
	if enabled, err := cast.ToBoolE(mode); err == nil && !enabled {
		c.Validation = ValidationModeDisabled
		return
	}
	if mode == "warnings" {
		c.Validation = ValidationModeWarnings
		return
	}
	// Our default is the strict mode.
	c.Validation = ValidationModeStrict
}
