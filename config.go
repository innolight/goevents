package goevents

import "time"

type writeGroupConfig struct {
	batchSize    int
	batchTimeout time.Duration
	workerCount  int
	bufferSize   int
	middlewares  []MiddleWare
}

const (
	DefaultPublisherCount        = 16
	DefaultSendChannelBufferSize = 1024
)

type Option = func(c *writeGroupConfig)

func WithPublisherCount(count int) Option {
	return func(c *writeGroupConfig) {
		c.workerCount = count
	}
}

func WithBufferSize(bufferSize int) Option {
	return func(c *writeGroupConfig) {
		c.bufferSize = bufferSize
	}
}

func WithMiddlewares(middlewares ...MiddleWare) Option {
	return func(c *writeGroupConfig) {
		c.middlewares = append(c.middlewares, middlewares...)
	}
}
