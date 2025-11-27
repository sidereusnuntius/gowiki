package queue

import (
	"time"

	"github.com/mikestefanello/backlite"
)

const (
	FetchQueue = "Fetch"
)

type FetchJob struct {
	Iri string
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