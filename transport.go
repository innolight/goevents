package goevents

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Transport interface {
	Send(topic string) (chan<- EventEnvelop, error)
	Receive(topic string) (<-chan EventEnvelop, error)
	Shutdown(ctx context.Context)
}

func NewTransport(queueFactory QueueProvider, options ...Option) Transport {
	var conf = writeGroupConfig{
		workerCount: DefaultPublisherCount,
		bufferSize:  DefaultSendChannelBufferSize,
	}
	for _, apply := range options {
		apply(&conf)
	}
	return &transport{
		conf:         conf,
		queueFactory: queueFactory,
	}
}

type transport struct {
	conf         writeGroupConfig
	queueFactory QueueProvider
	// map: topic -> writeGroup
	writeGroups sync.Map
}

func (t *transport) Send(topic string) (chan<- EventEnvelop, error) {
	val, ok := t.writeGroups.Load(topic)
	if ok {
		return val.(*writeGroup).GetChannel()
	}

	queue, err := t.getQueue(topic)
	if err != nil {
		return nil, err
	}

	channel := make(chan EventEnvelop, t.conf.bufferSize)
	writeGroup, err := newWriteGroup(channel, queue)
	go writeGroup.start(t.conf.workerCount)
	if err != nil {
		return nil, fmt.Errorf("NewSqsPublisher: %w", err)
	}

	t.writeGroups.Store(topic, writeGroup)
	return writeGroup.GetChannel()
}

func (t *transport) Receive(topic string) (<-chan EventEnvelop, error) {
	queue, err := t.getQueue(topic)
	if err != nil {
		return nil, err
	}
	if queue == nil {
		panic("shit")
	}
	result := make(chan EventEnvelop)
	go func() {
		for {
			events, err := queue.Receive(context.TODO())
			if err != nil {
				log.Printf("topic.Dequeue err%v\n", err)
				// TODO: use better backoff strategy
				time.Sleep(time.Second)
				continue
			}
			for _, event := range events {
				result <- event
			}
		}
	}()

	return result, nil
}

func (t *transport) getQueue(topic string) (Queue, error) {
	queue, err := t.queueFactory.Get(context.Background(), topic)
	if err != nil {
		return nil, fmt.Errorf("NewSqsPublisher: %w", err)
	}
	for i := len(t.conf.middlewares) - 1; i >= 0; i-- {
		middleware := t.conf.middlewares[i]
		queue = middleware(queue)
	}
	return queue, nil
}

func (t *transport) Shutdown(ctx context.Context) {
	start := time.Now()
	var waitGroup sync.WaitGroup
	t.writeGroups.Range(func(key, value interface{}) bool {
		waitGroup.Add(1)
		go func() {
			value.(*writeGroup).shutdown(ctx)
			waitGroup.Done()
		}()

		return true // return true to continue the iteration
	})
	waitGroup.Wait()
	log.Printf("[goevents] transport shutdown completed after %s\n", time.Since(start))
}

func newWriteGroup(
	backlog chan EventEnvelop,
	queue Queue,
) (*writeGroup, error) {
	ctx, cancel := context.WithCancel(context.Background())

	writeGroup := &writeGroup{
		topic:                 queue,
		backlog:               backlog,
		waitGroup:             sync.WaitGroup{},
		handlingCtx:           ctx,
		cancelHandlingContext: cancel,
	}

	return writeGroup, nil
}

// writeGroup represents a group of gorountines to write to the topic
type writeGroup struct {
	conf    writeGroupConfig
	topic   Queue
	backlog chan EventEnvelop

	// waitGroup to coordinate shutdown between all topic
	waitGroup sync.WaitGroup

	// context of the handling topic
	// context can be cancelled using cancelHandlingContext to force topic to abort sending messages
	handlingCtx           context.Context
	cancelHandlingContext func()
}

func (s *writeGroup) GetChannel() (chan<- EventEnvelop, error) {
	return s.backlog, nil
}

func (s *writeGroup) shutdown(ctx context.Context) {
	close(s.backlog)
	done := make(chan struct{})
	go func() {
		s.waitGroup.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		// Sending context cancellation signal
		log.Println("Shutdown context cancelled, cancelling pending sends")
		s.cancelHandlingContext()
		<-done
	case <-done:
		// all is good here
	}
}

func (s *writeGroup) start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		s.waitGroup.Add(1)
		go func() {
			defer s.waitGroup.Done()
			for message := range s.backlog {
				err := s.topic.Send(s.handlingCtx, message.Event)
				if message.Result != nil {
					message.Result <- err
				}
			}
		}()
	}
}

func (s *writeGroup) startWithBatching() {
	for i := 0; i < s.conf.workerCount; i++ {
		s.waitGroup.Add(1)
		go func() {

			defer s.waitGroup.Done()
			batch := make([]EventEnvelop, 0, s.conf.batchSize)
			batchTimeout := time.NewTicker(s.conf.batchTimeout)

			sendBatch := func() {
				batchCopy := batch
				s.sendBatch(batchCopy)
				batch = make([]EventEnvelop, 0, s.conf.batchSize)
				batchTimeout.Stop()
				batchTimeout = time.NewTicker(s.conf.batchTimeout)
			}

			for {
				select {
				case <-batchTimeout.C:
					sendBatch()
				case message, closed := <-s.backlog:
					if closed { // channel is closed (would yield immediately)
						sendBatch()
						return
					}
					batch = append(batch, message)
					if len(batch) == s.conf.batchSize {
						sendBatch()
					}
				}
			}
		}()
	}
}

func (s *writeGroup) sendBatch(batch []EventEnvelop) {
	if len(batch) == 0 {
		return
	}
	panic("add error handling")
}
