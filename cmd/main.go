package main

import (
	"k8s.io/klog/v2"
	"os"
	"zeus/cmd/commands"
	_ "zeus/docs"
)

// @title Zeus API
// @version 1.0
// @description kubernetes 多集群管理平台 API

// @BasePath /api/v1
func main() {
	cmd := commands.RootCommand()
	if err := cmd.Execute(); err != nil {
		klog.Fatalln(err)
		os.Exit(1)
	}
}
