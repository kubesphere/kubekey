package ending

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type TaskResult struct {
	mu            sync.Mutex
	ActionResults []ActionResult
	Status        ResultStatus
	StartTime     time.Time
	EndTime       time.Time
}

type ActionResult struct {
	Host   connector.Host
	Status ResultStatus
	Error  error
}

func NewTaskResult() *TaskResult {
	return &TaskResult{ActionResults: make([]ActionResult, 0, 0), Status: NULL, StartTime: time.Now()}
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

func (t *TaskResult) AppendSkip(host connector.Host) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e := ActionResult{
		Host:   host,
		Status: SKIPPED,
		Error:  nil,
	}

	t.ActionResults = append(t.ActionResults, e)
	t.EndTime = time.Now()
}

func (t *TaskResult) AppendSuccess(host connector.Host) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e := ActionResult{
		Host:   host,
		Status: SUCCESS,
		Error:  nil,
	}

	t.ActionResults = append(t.ActionResults, e)
	t.EndTime = time.Now()
}

func (t *TaskResult) AppendErr(host connector.Host, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e := ActionResult{
		Host:   host,
		Status: FAILED,
		Error:  err,
	}

	t.ActionResults = append(t.ActionResults, e)
	t.EndTime = time.Now()
	t.Status = FAILED
}

func (t *TaskResult) IsFailed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == FAILED {
		return true
	}
	return false
}

func (t *TaskResult) CombineErr() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.ActionResults) != 0 {
		var str string
		for i := range t.ActionResults {
			if t.ActionResults[i].Status != FAILED {
				continue
			}
			str += fmt.Sprintf("\nfailed: [%s] %s", t.ActionResults[i].Host.GetName(), t.ActionResults[i].Error.Error())
		}
		return errors.New(str)
	}
	return nil
}
