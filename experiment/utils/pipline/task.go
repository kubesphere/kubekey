package pipline

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"time"
)

type Task struct {
	Name        string
	Hosts       []kubekeyapiv1alpha1.HostCfg
	Action      Action
	Env         []map[string]string
	Vars        []map[string]interface{}
	Result      Result
	tag         string
	Parallel    bool
	Condition   bool
	IgnoreError bool
	Retry       int
	Delay       time.Time
	Serial      string
}

type Result struct {
	ResultCode int    // 0 or 1
	Status     string // success or failed
	Stdout     string
	Stderr     string
	StartTime  time.Time
	EndTime    time.Time
}

func (t *Task) Execute(vars *Vars) (Result, error) {
	if t.Parallel {
		go func() {}()
	} else {

	}
	return Result{}, nil
}



func (t *Task) Prepare() () {

}