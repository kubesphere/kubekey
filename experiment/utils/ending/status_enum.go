package ending

type ResultStatus int

const (
	NULL    ResultStatus = -99
	SKIPPED ResultStatus = iota - 1
	SUCCESS
	FAILED
)

var EnumList = []ResultStatus{
	NULL,
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
	case NULL:
		return "null"
	default:
		return "invalid option"
	}
}

func GetByCode(code int) ResultStatus {
	switch code {
	case -99:
		return NULL
	case -1:
		return SKIPPED
	case 0:
		return SUCCESS
	default:
		return FAILED
	}
}
