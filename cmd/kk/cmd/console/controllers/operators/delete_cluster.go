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

func DeleteCluster(c *gin.Context, targetDir string, tmpDir string) {
	//  升级连接
	clusterName := c.DefaultQuery("clusterName", "")
	deleteCRI := c.DefaultQuery("deleteCRI", "no")
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
			fmt.Println("删除集群失败")
			break
		}
		go func(message []byte) {
			// 写入文件
			mkdirErr := os.MkdirAll(fmt.Sprintf("./%s/%s", tmpDir, clusterName), 0755)
			if mkdirErr != nil {
				fmt.Println("创建目录时出错：", err)
				fmt.Println("删除集群失败")
				return
			}
			//fmt.Println("ClusterName is:::::", clusterName)
			filePath := fmt.Sprintf("./%s/%s/Cluster-deleteCluster.yaml", tmpDir, clusterName)
			err := os.WriteFile(filePath, readMsg, 0644)
			if err != nil {
				fmt.Println("写入文件时出错：", err)
				fmt.Println("删除集群失败")
				return
			}
			// 解析yaml数据到data
			var data kubekeyapiv1alpha2.Cluster
			unmarshalErr := yaml.Unmarshal(readMsg, &data)
			if unmarshalErr != nil {
				fmt.Println("websocket解析yaml出错:", unmarshalErr)
				fmt.Println("删除集群失败")
				return
			}
			arg := common.Argument{
				FilePath:          filePath,
				Debug:             false,
				KubernetesVersion: "",
				DeleteCRI:         deleteCRI == "yes",
				SkipConfirmCheck:  true,
			}
			actionErr := pipelines.DeleteCluster(arg)
			if actionErr != nil {
				msg := console_common.FormatErrorMessage(actionErr)
				fmt.Println(msg)
				fmt.Println("删除集群失败")
			} else {
				deleteFilePath := fmt.Sprintf("./%s/%s/Cluster.yaml", targetDir, clusterName)
				// 检查文件是否存在
				if _, err := os.Stat(deleteFilePath); err == nil {
					// 文件存在，进行删除操作
					err := os.Remove(deleteFilePath)
					if err != nil {
						fmt.Println("删除文件时出错：", err)
						return
					}
				} else if os.IsNotExist(err) {
					fmt.Println("文件不存在，不需要删除")
				} else {
					// 其他错误
					fmt.Println("检查文件时出错：", err)
					return
				}
				fmt.Println("删除集群成功")
			}
		}(readMsg)
	}
}
