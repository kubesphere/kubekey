package ending

import (
	"github.com/pkg/errors"
	"sync"
	"time"
)

type TaskResult struct {
	mu        sync.Mutex
	Endings   map[string]Ending
	Errors    []error
	Status    ResultStatus
	StartTime time.Time
	EndTime   time.Time
}

func NewTaskResult() *TaskResult {
	return &TaskResult{Endings: nil, Status: NULL, StartTime: time.Now()}
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
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Errors = append(t.Errors, err)
	t.EndTime = time.Now()
	t.Status = FAILED
}

func (t *TaskResult) AppendEnding(ending Ending, nodeName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Endings[nodeName] = ending
}

func (t *TaskResult) GetEnding(nodeName string) Ending {
	t.mu.Lock()
	defer t.mu.Unlock()
	ending := t.Endings[nodeName]
	return ending
}

func (t *TaskResult) StatisticEndings() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.Endings) == 0 {
		t.SkippedResult()
		return
	}

	for _, ending := range t.Endings {
		if ending.GetErr() != nil {
			t.AppendErr(ending.GetErr())
		}
		if ending.GetStatus() == FAILED {
			t.Status = FAILED
		}
	}

	if t.Status != FAILED && len(t.Errors) == 0 {
		t.NormalResult()
	}
}

func (t *TaskResult) CombineErr() error {
	if len(t.Errors) == 0 {
		return nil
	}
	var err error
	for i, v := range t.Errors {
		err = errors.Wrapf(v, "Error[%d]:", i)
	}
	return err
}

func (t *TaskResult) IsFailed() bool {
	if t.Status == FAILED {
		return true
	}
	return false
}

func (t *TaskResult) GetStatus() ResultStatus {
	if t.Status != NULL {
		return t.Status
	}
	t.StatisticEndings()
	return t.Status
}
