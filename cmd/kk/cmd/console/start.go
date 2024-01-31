package console

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/console/router"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/options"
	"github.com/spf13/cobra"
)

type ConsoleStartOptions struct {
	CommonOptions     *options.CommonOptions
	LogFilePath       string
	ConfigFileDirPath string
}

func NewConsoleStartOptions() *ConsoleStartOptions {
	return &ConsoleStartOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func NewCmdConsoleStart() *cobra.Command {
	o := NewConsoleStartOptions()
	// 这里的cmd只是声明，其中的Run字段要等输入了kk create cluster时才会运行
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a web console of kubekey",
		Run: func(cmd *cobra.Command, args []string) {
			go func() {
				fmt.Println(fmt.Sprintf("web控制台已启动，请访问[localhost:8082]或[公网ip:8082]查看"))
				err := openBrowser("localhost:8082")
				if err != nil {
				}
			}()
			ginServer := router.Router(o.LogFilePath, o.ConfigFileDirPath)
			err := ginServer.Run(":8082")
			if err != nil {
				log.Fatal("无法启动服务器:", err)
			} else {
			}
		},
	}
	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *ConsoleStartOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.LogFilePath, "log-path", "", "./server.log", "the path to the log of web server")
	cmd.Flags().StringVarP(&o.ConfigFileDirPath, "config-Path", "", "./config_from_console", "the path of the directory of config files")
}
