package manager

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

// defaultRetryBackoff is backoff with with duration of 5 seconds and factor of 2.0
func defaultRetryBackoff(retries int) wait.Backoff {
	return wait.Backoff{
		Steps:    retries,
		Duration: 5 * time.Second,
		Factor:   2.0,
	}
}

// Task is a runnable task
type Task struct {
	Fn      func(*Manager) error
	ErrMsg  string
	Retries int
}

// Run runs a task
func (t *Task) Run(mgrTask *Manager) error {
	if t.Retries == 0 {
		t.Retries = 1
	}
	backoff := defaultRetryBackoff(t.Retries)

	var lastError error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		if lastError != nil {
			mgrTask.Logger.Warn("Retrying task…")
		}
		lastError = t.Fn(mgrTask)
		if lastError != nil {
			mgrTask.Logger.Warn("Task failed…")
			if mgrTask.Verbose {
				mgrTask.Logger.Warnf("error was: %s", lastError)
			}
			return false, nil
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		err = lastError
	}
	return err
}
