package main

import (
	k8spipline "github.com/kubesphere/kubekey/experiment/kubernetes/piplines"
	"github.com/kubesphere/kubekey/experiment/utils/pipline"
)

func main() {
	groupVars := pipline.Vars{}
	err := k8spipline.CreateClusterPipeline.Start(&groupVars)
	if err != nil {
		return
	}
}
