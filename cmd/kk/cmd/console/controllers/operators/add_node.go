package operators

import (
	"fmt"
	"github.com/gin-gonic/gin"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/console/console_common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/pipelines"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func AddNode(c *gin.Context, targetDir string, tmpDir string) {
	//  升级连接
	clusterName := c.DefaultQuery("clusterName", "")
	clientConn, err := console_common.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatalf("Failed to set websocket upgrade: %v", err)
		return
	}
	// 建立管道stdout->websocket，监听信息
	captureBuffer := console_common.SetupCaptureBuffer()
	captureBuffer.WsConn = clientConn
	for {
		_, readMsg, readErr := clientConn.ReadMessage()

		if readErr != nil {
			fmt.Println("websocket后台读取消息出错:", err)
			fmt.Println("添加节点失败")
			break
		}
		go func(message []byte) {
			// 写入文件
			mkdirErr := os.MkdirAll(fmt.Sprintf("./%s/%s", tmpDir, clusterName), 0755)
			if mkdirErr != nil {
				fmt.Println("创建目录时出错：", err)
				fmt.Println("添加节点失败")
				return
			}
			filePath := fmt.Sprintf("./%s/%s/Cluster-addNode.yaml", tmpDir, clusterName)
			err := os.WriteFile(filePath, readMsg, 0644)
			if err != nil {
				fmt.Println("写入文件时出错：", err)
				fmt.Println("添加节点失败")
				return
			}
			// 解析yaml数据到data
			var data kubekeyapiv1alpha2.Cluster
			unmarshalErr := yaml.Unmarshal(readMsg, &data)
			if unmarshalErr != nil {
				fmt.Println("websocket解析yaml出错:", unmarshalErr)
				fmt.Println("添加节点失败")
				return
			}
			arg := common.Argument{
				FilePath:         filePath,
				Debug:            false,
				IgnoreErr:        false,
				SkipConfirmCheck: true,
				SkipPullImages:   false,
				ContainerManager: data.Spec.Kubernetes.ContainerManager,
				Artifact:         "",
				InstallPackages:  false,
			}
			actionErr := pipelines.AddNodes(arg, "")
			if actionErr != nil {
				msg := console_common.FormatErrorMessage(actionErr)
				fmt.Println(msg)
				fmt.Println("添加节点失败")
			} else {
				fmt.Println("添加节点成功")
				mkdirErr := os.MkdirAll(fmt.Sprintf("./%s/%s", targetDir, clusterName), 0755)
				if mkdirErr != nil {
					fmt.Println("创建目录时出错：", err)
					return
				}
				filePath := fmt.Sprintf("./%s/%s/Cluster.yaml", targetDir, clusterName)
				err := os.WriteFile(filePath, readMsg, 0644)
				if err != nil {
					fmt.Println("写入文件时出错：", err)
					return
				}
			}
		}(readMsg)
	}

}
