package ending

import (
	"github.com/kubesphere/kubekey/pkg/core/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"time"
)

type ModuleResult struct {
	HostResults   map[string]Interface
	CombineResult error
	Status        ResultStatus
	StartTime     time.Time
	EndTime       time.Time
}

func NewModuleResult() *ModuleResult {
	return &ModuleResult{HostResults: make(map[string]Interface), StartTime: time.Now(), Status: NULL}
}

func (m *ModuleResult) IsFailed() bool {
	if m.Status == FAILED {
		return true
	}
	return false
}

func (m *ModuleResult) AppendHostResult(p Interface) {
	if m.HostResults == nil {
		return
	}
	m.HostResults[p.GetHost().GetName()] = p
}

func (m *ModuleResult) LocalErrResult(err error) {
	now := time.Now()
	r := &ActionResult{
		Host:      &connector.BaseHost{Name: common.LocalHost},
		Status:    FAILED,
		Error:     err,
		StartTime: m.StartTime,
		EndTime:   now,
	}

	m.HostResults[r.Host.GetName()] = r
	m.CombineResult = err
	m.EndTime = now
	m.Status = FAILED
}

func (m *ModuleResult) ErrResult(combineErr error) {
	m.EndTime = time.Now()
	m.Status = FAILED
	m.CombineResult = combineErr
}

func (m *ModuleResult) NormalResult() {
	m.EndTime = time.Now()
	m.Status = SUCCESS
}
