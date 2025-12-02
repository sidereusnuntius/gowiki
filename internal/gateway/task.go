package gateway

import (
	"time"

	"github.com/mikestefanello/backlite"
)

type TaskType uint8

const (
	Fetch TaskType = iota
	Deliver
)

type Task struct {
	Type    TaskType
	To      string
	From    string
	Payload map[string]any
	Next    *Task
}

func (t Task) Config() backlite.QueueConfig {
	return backlite.QueueConfig{
		Name:        "tasks",
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
