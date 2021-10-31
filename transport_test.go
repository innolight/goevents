package goevents

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTransport_GracefulShutdownOnTime(t *testing.T) {
	// GIVEN
	publisherFactory := &mockPublisherFactory{
		publishLatency: func(event Event) time.Duration { return time.Millisecond * 10 },
		publishError:   func(ctx context.Context, event Event) error { return nil },
	}
	transport := &transport{
		conf: writeGroupConfig{
			workerCount: 10,
			bufferSize:  1024,
		},
		queueFactory: publisherFactory,
	}
	assert.NotNil(t, transport)

	// WHEN send events and shutdowns
	sendChan, err := transport.Send("cool-topic")
	if err != nil {
		t.FailNow()
	}

	publishedCount, result := publishFor(time.Millisecond*100, sendChan)
	var errs []error
	go func() {
		for r := range result {
			if r != nil {
				errs = append(errs, err)
			}
		}
	}()

	transport.Shutdown(context.Background())
	close(result)

	// THEN
	assert.Equal(t, int(publishedCount), len(publisherFactory.GetPublishedEvents()))
	assert.Equal(t, 0, len(errs))
}

func TestTransport_GracefulShutdownTimedOut(t *testing.T) {
	// GIVEN
	publisherFactory := &mockPublisherFactory{
		publishLatency: func(event Event) time.Duration { return time.Second },
		publishError: func(ctx context.Context, event Event) error {
			return ctx.Err()
		},
	}
	transport := &transport{
		conf: writeGroupConfig{
			workerCount: 10,
			bufferSize:  1024,
		},
		queueFactory: publisherFactory,
	}
	assert.NotNil(t, transport)

	sendChan, err := transport.Send("cool-topic")
	if err != nil {
		t.FailNow()
	}

	// WHEN
	publishedCount, resultChan := publishFor(time.Millisecond*10, sendChan)

	// shut down in an already cancelled context
	shutDownContext, cancelCtx := context.WithCancel(context.Background())
	var errs []error
	go func() {
		for err := range resultChan {
			if err != nil {
				errs = append(errs, err)
			}
		}
	}()
	go func() {
		time.Sleep(time.Millisecond * 100)
		cancelCtx()
	}()
	transport.Shutdown(shutDownContext)
	close(resultChan)

	assert.NotEmpty(t, publishedCount)
	assert.NotEmpty(t, errs)
	assert.Equal(t, int(publishedCount), len(publisherFactory.GetPublishedEvents())+len(errs))
	fmt.Printf("err %d/%d\n", len(errs), publishedCount)
}

func publishFor(duration time.Duration, sendChan chan<- EventEnvelop) (publishedCount uint64, result chan error) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	var writeGroup sync.WaitGroup
	resultChan := make(chan error)
	for i := 0; i < 100; i++ {
		idx := i
		writeGroup.Add(1)
		go func() {
			defer writeGroup.Done()
		loop:
			for {
				event := EventEnvelop{
					Result: resultChan,
					Event:  StringEvent(context.Background(), fmt.Sprintf("p%d %d", idx, time.Now().UnixNano())),
				}
				select {
				case <-ctx.Done():
					break loop
				case sendChan <- event:
					atomic.AddUint64(&publishedCount, 1)
				}
			}
		}()
	}
	time.Sleep(duration)
	cancelCtx()
	writeGroup.Wait()
	fmt.Println("done publishing")
	return publishedCount, resultChan
}

type mockPublisherFactory struct {
	events         []Event
	mut            sync.Mutex
	publishLatency func(event Event) time.Duration
	publishError   func(ctx context.Context, event Event) error
}

func (m *mockPublisherFactory) Get(ctx context.Context, name string) (Queue, error) {
	return &mockPublisher{
		publish: func(ctx context.Context, event Event) error {
			if err := m.publishError(ctx, event); err != nil {
				//loggingWrapper.Globalf("publishing err %s", event.Body)
				return err
			}

			time.Sleep(m.publishLatency(event))
			m.mut.Lock()
			m.events = append(m.events, event)
			m.mut.Unlock()
			log.Printf("published %s\n", event.Body)
			return nil
		},
		publishBatch: func(ctx context.Context, events []Event) error {
			// not used
			return nil
		},
	}, nil
}

func (m *mockPublisherFactory) GetPublishedEvents() []Event {
	m.mut.Lock()
	defer m.mut.Unlock()
	return m.events
}

type mockPublisher struct {
	publish      func(ctx context.Context, messageCarrier Event) error
	publishBatch func(ctx context.Context, messageCarriers []Event) error
}

func (m mockPublisher) Send(ctx context.Context, messageCarrier Event) error {
	return m.publish(ctx, messageCarrier)
}

func (m mockPublisher) Receive(ctx context.Context) ([]EventEnvelop, error) {
	panic("implement me")
}
