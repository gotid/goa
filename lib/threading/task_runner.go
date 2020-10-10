package threading

import (
	"github.com/z-sdk/goa/lib/lang"
	"github.com/z-sdk/goa/lib/rescue"
)

type TaskRunner struct {
	limitChan chan lang.PlaceholderType
}

func (r *TaskRunner) Schedule(taskFn func()) {
	r.limitChan <- lang.Placeholder

	go func() {
		defer rescue.Recover(func() {
			<-r.limitChan
		})

		taskFn()
	}()
}

func NewTaskRunner(concurrency int) *TaskRunner {
	return &TaskRunner{limitChan: make(chan lang.PlaceholderType, concurrency)}
}
