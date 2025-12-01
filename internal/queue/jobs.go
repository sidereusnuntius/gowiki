package queue

import (
	"time"

	"github.com/mikestefanello/backlite"
)

const (
	FetchQueue    = "Fetch"
	DeliveryQueue = "Delivery"
)

type FetchJob struct {
	Iri  string
	Next *PostJob
}

func (j FetchJob) Config() backlite.QueueConfig {
	return backlite.QueueConfig{
		Name:        FetchQueue,
		MaxAttempts: 5,
		Backoff:     5 * time.Second,
		Timeout:     10 * time.Second,
		Retention: &backlite.Retention{
			Duration:   12 * time.Hour,
			OnlyFailed: false,
			Data: &backlite.RetainData{
				OnlyFailed: true,
			},
		},
	}
}

type PostJob struct {
	To   string
	From string
	Body map[string]any
	Next backlite.Task
}

func (j PostJob) Config() backlite.QueueConfig {
	return backlite.QueueConfig{
		Name:        DeliveryQueue,
		MaxAttempts: 5,
		Backoff:     5 * time.Second,
		Timeout:     10 * time.Second,
		Retention: &backlite.Retention{
			Duration:   12 * time.Hour,
			OnlyFailed: false,
			Data: &backlite.RetainData{
				OnlyFailed: true,
			},
		},
	}
}
