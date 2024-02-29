package clusterinfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/utils"
	"github.com/xuri/excelize/v2"
)

type DumpOption struct {
	Namespace     []string
	KubeConfig    string
	AllNamespaces bool
	OutputDir     string
	Type          string
	Tar           bool
	Queue         int
	Logger        bool
}

func Dump(option DumpOption) error {
	client, err := utils.NewClient(option.KubeConfig)
	if err != nil {
		return err
	}
	dump := NewDumpOption(client)

	cluster, err := dump.GetMultiCluster()
	if err != nil {
		return err
	}

	fileChan := NewFileChan(option, excelize.NewFile())
	queue := make(chan struct{}, option.Queue)
	go func() {
		err = fileChan.WriteFile()
		if err != nil {
			fmt.Printf("failed to write file %s", err.Error())
			fileChan.WaitGroup.Done()
		}
	}()

	for _, multiCluster := range cluster {
		fileChan.WaitGroup.Add(1)
		queue <- struct{}{}
		go func(multiCluster v1alpha2.MultiCluster) {
			defer func() {
				<-queue
			}()
			if multiCluster.IsHostCluster() {
				fileChan.ReadData(dump, multiCluster.Name)
			} else {
				clusterClient, err := utils.NewClientForCluster(multiCluster.Spec.Connection.KubeConfig)
				if err != nil {
					fmt.Printf("failed to create cluster %s", multiCluster.Name)
					fileChan.WaitGroup.Done()
					return
				}
				fileChan.ReadData(NewDumpOption(clusterClient), multiCluster.Name)
			}
		}(multiCluster)
	}

	defer func() {
		if err := fileChan.Excel.Close(); err != nil {
			fmt.Println(err)
		}
		close(fileChan.OutFileChan)
		close(queue)
	}()

	fileChan.WaitGroup.Wait()

	fileChan.Excel.DeleteSheet("Sheet1")
	fileChan.Excel.SaveAs(fmt.Sprintf("%s/%s", option.GetOutputDir(), "cluster_dump.xlsx"))

	if option.Tar {
		err = NewTar(option).Run()
		if err != nil {
			fmt.Printf("failed to tar file %s", err.Error())
			return err
		}
	}

	return nil
}

func resourcesClassification(resources interface{}) map[string][]interface{} {

	var resourcesMap []map[string]interface{}
	if marshal, err := json.Marshal(resources); err != nil {
		fmt.Println(err, "marshal resources error")
		return nil
	} else {
		decoder := json.NewDecoder(bytes.NewReader(marshal))
		if err = decoder.Decode(&resourcesMap); err != nil {
			fmt.Println(err, "Decode resources error")
			return nil
		}
	}

	var completeMap = make(map[string][]interface{})
	for _, m := range resourcesMap {
		namespace, ok := m["metadata"].(map[string]interface{})["namespace"]
		if ok {
			completeMap[namespace.(string)] = append(completeMap[namespace.(string)], m)
		} else {
			completeMap[""] = append(completeMap[""], m)
		}
	}
	return completeMap
}

func (c *DumpOption) GetOutputDir() string {
	if c.OutputDir == "" {
		return "cluster_dump"
	}
	return c.OutputDir
}
