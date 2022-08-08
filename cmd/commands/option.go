package commands

import (
	"flag"
	"k8s.io/client-go/util/homedir"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"path/filepath"
)

const (
	defaultPort                  = 8080
	defaultConfigPath            = "config.yaml"
	defaultMaxConnectionLifetime = 10
	defaultMaxConnectionIdleTime = 60
	defaultAdminClusterName      = "admin-cluster"
)

type commandOptions struct {
	Port             int
	Debug            bool
	ConfigFile       string
	AdminKubeConfig  string
	AdminClusterName string
}

func newCommandOptions() *commandOptions {
	return &commandOptions{
		Port:       0,
		ConfigFile: "",
	}
}

func (opts *commandOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("zeus")
	fs.IntVarP(&opts.Port, "port", "", defaultPort, "http server listen port.")
	fs.BoolVarP(&opts.Debug, "debug", "", false, "http server debug mode.")
	fs.StringVarP(&opts.ConfigFile, "config", "", defaultConfigPath, "config fileutil path.")
	fs.StringVarP(&opts.AdminClusterName, "admin-cluster-name", "", defaultAdminClusterName, "if not set ")
	fs.StringVarP(&opts.AdminKubeConfig, "kubeconfig", "", filepath.Join(homedir.HomeDir(), ".kube", "config"), "kubeconfig fileutil path.")
	klogfs := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogfs)
	fss.FlagSet("klog").AddGoFlagSet(klogfs)
	return fss
}
