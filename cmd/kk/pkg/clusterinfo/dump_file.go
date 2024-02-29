package clusterinfo

import (
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/strings/slices"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"sync"
)

type ChanData struct {
	FilePath string
	Data     interface{}
}
type DumpFileChan struct {
	DumpOption
	OutFileChan chan ChanData
	Excel       *excelize.File
	WaitGroup   sync.WaitGroup
}

type DumpFile interface {
	CreateFile(string) (*os.File, error)
	GetWireData(ChanData) []byte
	WriteFile() error
	SendData(ChanData, bool)
	ReadData(*ClientSet, string)
}

func NewFileChan(option DumpOption, excel *excelize.File) *DumpFileChan {
	return &DumpFileChan{
		OutFileChan: make(chan ChanData),
		WaitGroup:   sync.WaitGroup{},
		DumpOption:  option,
		Excel:       excel,
	}
}

func (c *DumpFileChan) CreateFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0775); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (c *DumpFileChan) GetWireData(o ChanData) []byte {
	switch strings.ToLower(c.Type) {
	case "yaml":
		marshal, err := yaml.Marshal(o.Data)
		if err != nil {
			fmt.Printf("marshal data error %s\n", err.Error())
			return nil
		}
		return marshal
	default:
		marshal, err := json.Marshal(o.Data)
		if err != nil {
			fmt.Printf("marshal data error %s\n", err.Error())
			return nil
		}
		return marshal
	}
}

func (c *DumpFileChan) WriteFile() error {

	for o := range c.OutFileChan {
		file, err := c.CreateFile(o.FilePath)
		if err != nil {
			fmt.Println(err, "create file error")
			c.WaitGroup.Done()
			return err
		}
		data := c.GetWireData(o)
		if _, err = file.Write(data); err != nil {
			fmt.Println(err, "write file error")
			c.WaitGroup.Done()
			return err
		}
		if c.Logger {
			fmt.Println(string(data))
		}
	}
	return nil
}

func (c *DumpFileChan) SendData(data ChanData, isContinue bool) {
	if isContinue || c.AllNamespaces {
		c.OutFileChan <- data
	}
}

func (c *DumpFileChan) ReadData(client *ClientSet, clusterName string) {

	_, err := c.Excel.NewSheet(clusterName)
	if err != nil {
		fmt.Println(err, "create sheet error")
		return
	}

	index, err := c.Excel.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 15,
		},
	})
	resources := client.GetClusterResources(corev1.NamespaceAll)
	cellTag := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	cellTagIndex := 0
	for resourceKey, resource := range resources {
		for kind, instance := range resource {
			c.Excel.SetCellValue(clusterName, fmt.Sprintf("%s1", cellTag[cellTagIndex]), kind)
			rowIndex := 2
			for namespace, data := range instance {
				for _, datum := range data {
					name := datum.(map[string]interface{})["metadata"].(map[string]interface{})["name"]
					c.Excel.SetCellValue(clusterName, fmt.Sprintf("%s%d", cellTag[cellTagIndex], rowIndex), name)
					c.SendData(ChanData{
						FilePath: filepath.Join(c.GetOutputDir(), clusterName, resourceKey, namespace, kind, fmt.Sprintf("%s.%s", name, strings.ToLower(c.Type))),
						Data:     datum,
					}, slices.Contains(c.Namespace, namespace) || namespace == "")
					rowIndex++
				}
			}
			cellTagIndex++
		}
	}

	c.Excel.SetCellStyle(clusterName, "A1", fmt.Sprintf("%s1", cellTag[cellTagIndex-1]), index)
	c.Excel.SetColWidth(clusterName, "A", fmt.Sprintf("%s", cellTag[cellTagIndex-1]), 30)
	c.WaitGroup.Done()

}
