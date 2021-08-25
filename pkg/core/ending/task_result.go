package ending

import (
	"time"
)

type TaskResult struct {
	Error     error
	Status    ResultStatus
	StartTime time.Time
	EndTime   time.Time
}

func NewTaskResult() *TaskResult {
	return &TaskResult{Error: nil, Status: NULL, StartTime: time.Now()}
}

func (t *TaskResult) ErrResult() {
	if t.Status != NULL {
		return
	}
	t.EndTime = time.Now()
	t.Status = FAILED
}

func (t *TaskResult) NormalResult() {
	if t.Status != NULL {
		return
	}
	t.EndTime = time.Now()
	t.Status = SUCCESS
}

func (t *TaskResult) SkippedResult() {
	if t.Status != NULL {
		return
	}
	t.EndTime = time.Now()
	t.Status = SKIPPED
}

func (t *TaskResult) AppendErr(err error) {
	t.Error = err
	t.EndTime = time.Now()
	t.Status = FAILED
}

func (t *TaskResult) IsFailed() bool {
	if t.Status == FAILED {
		return true
	}
	return false
}
