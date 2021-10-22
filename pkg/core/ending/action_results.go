package ending

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"time"
)

type ActionResult struct {
	Host      connector.Host
	Status    ResultStatus
	Error     error
	StartTime time.Time
	EndTime   time.Time
}

func (a *ActionResult) GetHost() connector.Host {
	return a.Host
}

func (a *ActionResult) GetStatus() ResultStatus {
	return a.Status
}

func (a *ActionResult) GetErr() error {
	return a.Error
}

func (a *ActionResult) GetStartTime() time.Time {
	return a.StartTime
}

func (a *ActionResult) GetEndTime() time.Time {
	return a.EndTime
}
