package main

import (
	"flag"
	"fmt"
	"github.com/kubesphere/kubekey/v4/cmd/api-server/app"
	"os"
)

func main() {
	if err := app.ApiServerCommand().Execute(); err != nil {
		vFlag := flag.Lookup("v")
		if vFlag != nil {
			fmt.Printf("%+v", err)
		} else {
			fmt.Printf("%v", err)
		}
		os.Exit(1)
	}
}
