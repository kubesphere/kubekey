package router

import (
	"embed"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/console/console_common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/console/controllers/operators"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/version/kubernetes"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/version/kubesphere"
	"gopkg.in/yaml.v3"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

//go:embed templates/*
var embeddedTemplates embed.FS

//go:embed templates/static/*
var embeddedStatic embed.FS

func Router(LogFilePath string, ConfigFileDirPath string) *gin.Engine {
	fmt.Println(fmt.Sprintf("服务器日志存放地址：%s", LogFilePath))
	file, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件:", err)
	}
	defer file.Close()
	// 设置 gin 的日志输出为标准日志
	gin.DefaultWriter = file
	gin.SetMode(gin.ReleaseMode) // 设置 Gin 的模式为发布模式，以减少日志输出

	ginServer := gin.Default()
	staticFS, err := fs.Sub(embeddedStatic, "templates/static")
	if err != nil {
		// 处理错误
	}

	ginServer.StaticFS("/static", http.FS(staticFS))
	ginConfig := cors.DefaultConfig()
	ginConfig.AllowAllOrigins = true
	ginConfig.AllowMethods = []string{"GET", "POST"}
	ginConfig.AllowHeaders = []string{"Origin", "Content-Type"}
	ginServer.Use(cors.New(ginConfig))
	targetDir := ConfigFileDirPath
	tmpDir := "./tmp_config"

	ginServer.GET("/", func(context *gin.Context) {
		data, _ := embeddedTemplates.ReadFile("templates/index.html")
		context.Data(http.StatusOK, "text/html", data)
	})

	ginServer.GET("/scanCluster", func(c *gin.Context) {
		//fmt.Println("进入scanCluster")
		var clusterList []console_common.Cluster
		err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			relativePath, err := filepath.Rel(targetDir, path)
			if err != nil {
				return err
			}
			// 只检查一级子目录
			if info.IsDir() && filepath.Dir(relativePath) == "." && relativePath != "." {
				filePath := filepath.Join(path, "Cluster.yaml")
				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					// 文件存在
					fileContent, err := os.ReadFile(filePath)
					if err != nil {
						return err
					}
					var cluster console_common.Cluster
					if err := yaml.Unmarshal(fileContent, &cluster); err != nil {
						fmt.Println("Error parsing YAML:", err)
					} else {
						clusterList = append(clusterList, cluster)
					}
				}
				return filepath.SkipDir // 跳过此目录的其他文件和子目录
			}
			return nil
		})

		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"clusterData": clusterList})
	})

	ginServer.GET("/clusterVersionOptions", func(context *gin.Context) {
		//context.JSON 返回JSON
		context.JSON(200, gin.H{"clusterVersionOptions": kubernetes.SupportedK8sVersionList()})
	})

	ginServer.GET("/ksVersionOptions/:clusterVersion", func(context *gin.Context) {
		clusterVersion := context.Param("clusterVersion")
		if len(clusterVersion) < 5 {
			context.JSON(200, "invalid k8s version format")
		}
		compatibleKSVersions := []string{}
		k8sVersionMajorMinor := clusterVersion[:5]

		for _, installer := range kubesphere.VersionMap {
			for _, supportVersion := range installer.K8sSupportVersions {
				//fmt.Println(supportVersion)
				if supportVersion == k8sVersionMajorMinor {
					compatibleKSVersions = append(compatibleKSVersions, installer.Version)
					break
				}
			}
		}
		sort.Strings(compatibleKSVersions)
		sort.Sort(sort.Reverse(sort.StringSlice(compatibleKSVersions)))
		context.JSON(200, gin.H{"ksVersionOptions": compatibleKSVersions})
	})

	ginServer.GET("/createCluster", func(c *gin.Context) {
		operators.CreateCluster(c, targetDir, tmpDir)
	})

	ginServer.GET("/upgradeCluster", func(c *gin.Context) {
		operators.UpgradeCluster(c, targetDir, tmpDir)
	})
	ginServer.GET("/deleteNode", func(c *gin.Context) {
		operators.DeleteNode(c, targetDir, tmpDir)
	})

	ginServer.GET("/deleteCluster", func(c *gin.Context) {
		operators.DeleteCluster(c, targetDir, tmpDir)
	})

	ginServer.GET("/addNode", func(c *gin.Context) {
		operators.AddNode(c, targetDir, tmpDir)
	})
	return ginServer
}
