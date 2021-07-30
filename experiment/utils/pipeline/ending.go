package pipeline

import "time"

type Ending interface {
	GetStdOutString() string
	GetStatus() ResultStatus
	GetStartTime() time.Time
	GetEndTime() time.Time
	SetEndTime()
	GetErr() error
	ErrResult(err error)
}

type Result struct {
	Err        error
	ResultCode int    // 0 or 1
	Status     string // success or failed
	Stdout     string
	Stderr     string
	StartTime  time.Time
	EndTime    time.Time
}

func NewResult() *Result {
	return &Result{StartTime: time.Now()}
}

func NewResultWithErr(err error) *Result {
	return &Result{
		Err:        err,
		Stdout:     err.Error(),
		ResultCode: int(FAILED),
		StartTime:  time.Now(),
		EndTime:    time.Now(),
	}
}

func (r *Result) SetEndTime() {
	r.EndTime = time.Now()
}

func (r *Result) GetStartTime() time.Time {
	return r.StartTime
}

func (r *Result) GetEndTime() time.Time {
	return r.EndTime
}

func (r *Result) GetErr() error {
	return r.Err
}

func (r *Result) GetStdOutString() string {
	return r.Stdout
}

func (r *Result) GetStatus() ResultStatus {
	return GetByCode(r.ResultCode)
}

func (r *Result) ErrResult(err error) {
	r.Err = err
	r.Stdout = err.Error()
	r.Stderr = err.Error()
	r.ResultCode = int(FAILED)
}
