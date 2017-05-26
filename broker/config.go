package broker

import "errors"

type Config struct {
	QueueMain    string
	QueueInvalid string
	QueueError   string
}

func (c *Config) Validate() error {
	if c.QueueMain == "" {
		return errors.New("main queue name is undefined")
	}
	if c.QueueInvalid == "" {
		return errors.New("invalid queue name is undefined")
	}
	if c.QueueError == "" {
		return errors.New("error queue name is undefined")
	}
	return nil
}
