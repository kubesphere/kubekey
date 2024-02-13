package clusterinfo

import (
	"fmt"
	"os/exec"
)

const tarFileName = "cluster_dump.tar"

type TarOptions struct {
	DumpOption
}

func NewTar(option DumpOption) *TarOptions {
	return &TarOptions{
		DumpOption: option,
	}
}

func (t TarOptions) Run() error {

	sourceDir := t.GetOutputDir()
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("tar -czvf %s %s", tarFileName, sourceDir))
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	if t.Logger {
		fmt.Println(string(output))
	}

	return nil
}
