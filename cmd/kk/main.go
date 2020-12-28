package main

import (
	"github.com/kubesphere/kubekey/cmd"
)

// Using a separate entry-point can reduce the size of the binary file
func main() {
	cmd.Execute()
}
