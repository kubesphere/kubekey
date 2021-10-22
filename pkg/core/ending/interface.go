package ending

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"time"
)

type Interface interface {
	GetHost() connector.Host
	GetStatus() ResultStatus
	GetErr() error
	GetStartTime() time.Time
	GetEndTime() time.Time
}
