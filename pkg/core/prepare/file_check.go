package prepare

import "github.com/kubesphere/kubekey/pkg/core/connector"

type FileExist struct {
	BasePrepare
	FilePath string
	Not      bool
}

func (f *FileExist) PreCheck(runtime connector.Runtime) (bool, error) {
	exist, err := runtime.GetRunner().FileExist(f.FilePath)
	if err != nil {
		return false, err
	}
	if f.Not {
		return !exist, nil
	}
	return exist, nil
}
