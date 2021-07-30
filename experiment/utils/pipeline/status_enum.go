package pipeline

type ResultStatus int

const (
	SKIPPED ResultStatus = iota - 1
	SUCCESS
	FAILED
)

var EnumList = []ResultStatus{
	SKIPPED,
	SUCCESS,
	FAILED,
}

func (r ResultStatus) String() string {
	switch r {
	case SUCCESS:
		return "success"
	case FAILED:
		return "failed"
	case SKIPPED:
		return "skipped"
	default:
		return "invalid option"
	}
}

func GetByCode(code int) ResultStatus {
	switch code {
	case -1:
		return SKIPPED
	case 0:
		return SUCCESS
	default:
		return FAILED
	}
}
