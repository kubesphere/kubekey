package main

import (
	k8spipline "github.com/kubesphere/kubekey/experiment/kubernetes/piplines"
)

func main() {
	err := k8spipline.CreateClusterPipeline.Start()
	if err != nil {
		return
	}
}
