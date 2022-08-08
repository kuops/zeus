package commands

import (
	"context"
	"time"
	"zeus/config"
	"zeus/internal/handler"
	"zeus/internal/repository"
	"zeus/internal/server"
	"zeus/internal/service"
	"zeus/pkg/database"
	v1 "zeus/pkg/kubernetes/apis/cluster/v1"
	clusterclientset "zeus/pkg/kubernetes/client/clientset/versioned"
	clsuterinformers "zeus/pkg/kubernetes/client/informers/externalversions"
	"zeus/pkg/kubernetes/clusterresource"
	"zeus/pkg/kubernetes/controller/cluster"
	"zeus/pkg/signals"
	"zeus/pkg/util/fileutil"
	"zeus/pkg/util/kubeutil"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
	"k8s.io/klog/v2"
)

func RootCommand() *cobra.Command {
	opts := newCommandOptions()
	cmd := &cobra.Command{
		Use:           "zeus",
		Short:         "multi kubernetes cluster platform",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(opts)
		},
	}

	fs := cmd.Flags()
	namedFlagSets := opts.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cliflag.SetUsageAndHelpFunc(cmd, namedFlagSets, cols)

	return cmd
}

func Run(options *commandOptions) error {
	cfg := &config.Config{
		Server: &server.Config{
			Port:  options.Port,
			Debug: options.Debug,
		},
		Database: &database.Config{
			MaxConnectionLifetime: time.Second * defaultMaxConnectionLifetime,
			MaxConnectionIdleTime: time.Second * defaultMaxConnectionIdleTime,
		},
	}

	if err := cfg.ParseFromFile(options.ConfigFile); err != nil {
		klog.Warningf("Failed parse config file, using default config.")
	}
	stopCh := signals.SetupSignalHandler()
	clusterResources := clusterresource.NewClusterResources()
	clusterController := newClusterController(options, stopCh, clusterResources)
	clusterClientSet := clusterController.GetClusterClientSet()

	go func() {
		if err := clusterController.Run(2, stopCh); err != nil {
			klog.Fatalf("Error start cluster controller: %v", err)
		}
		<-stopCh
		for clusterName, cluster := range clusterResources {
			klog.Infof("close %v cluster informers", clusterName)
			close(cluster.StopCh)
		}
	}()

	repos := repository.NewRepositories(clusterClientSet, clusterResources)
	services := service.NewServices(repos)
	handlers := handler.NewHandlers(services)
	srv := server.New(cfg.Server)
	return srv.Run(handlers, stopCh)
}

func newClusterController(options *commandOptions, stopCh <-chan struct{},
	clusterResources map[string]*clusterresource.ClusterResource) *cluster.Controller {

	admRestConfig, err := kubeutil.GetRestConfig(options.AdminKubeConfig)
	if err != nil {
		klog.Fatalf("Error building admin cluster kubeConfig: %s", err.Error())
	}
	// admin cluster 管理集群 kubernetes clientSet
	admKubeClientSet, err := kubernetes.NewForConfig(admRestConfig)
	if err != nil {
		klog.Fatalf("Error building admin cluster kubeClientSet: %s", err.Error())
	}
	// code-generator 生成的 clusters.cluster.shiny.io clientSet
	admClusterClientSet, err := clusterclientset.NewForConfig(admRestConfig)
	if err != nil {
		klog.Fatalf("Error building admin cluster clusterClientSet: %s", err.Error())
	}
	// 检查集群健康
	err = admClusterClientSet.Discovery().RESTClient().Get().AbsPath("/healthz").Do(context.TODO()).Error()
	if err != nil {
		klog.Fatalf("Error connect admin cluster, %v", err)
	}
	// 添加 adminCluster CRD
	admKubeConfig, err := fileutil.FileContent(options.AdminKubeConfig)
	if err != nil {
		klog.Errorf("Error read admin cluster kubeConfig fileutil, %v", err)
	}
	admCluster, err := admClusterClientSet.ClusterV1().Clusters().Get(context.TODO(), options.AdminClusterName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = admClusterClientSet.ClusterV1().Clusters().Create(context.TODO(), &v1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: options.AdminClusterName,
			},
			Spec: v1.ClusterSpec{
				KubeConfig: admKubeConfig,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("Error create admin cluster resources, %#v", err)
		}
	} else {
		if options.AdminClusterName != defaultAdminClusterName {
			if _, err := admClusterClientSet.ClusterV1().Clusters().Get(context.TODO(), defaultAdminClusterName, metav1.GetOptions{}); err == nil {
				_ = admClusterClientSet.ClusterV1().Clusters().Delete(context.TODO(), defaultAdminClusterName, metav1.DeleteOptions{})
			}
		}
		if admCluster.Spec.KubeConfig != admKubeConfig {
			admCluster.Spec.KubeConfig = admKubeConfig
			_, err = admClusterClientSet.ClusterV1().Clusters().Update(context.TODO(), admCluster, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("Error update admin cluster resources, %#v", err)
			}
		}
	}
	// admClusterInformerFactory clusters.cluster.shiny.io 的 informerFactory
	admClusterInformerFactory := clsuterinformers.NewSharedInformerFactory(admClusterClientSet, time.Minute)
	// clusterController 初始化 Controller ，创建 clusterInformer 实例
	clusterController := cluster.NewController(
		admKubeClientSet, admClusterClientSet,
		admClusterInformerFactory.Cluster().V1().Clusters(),
		clusterResources)
	// informerFactory 创建具体 informer 实例之后，启动 informerFactory 中所有 informer
	admClusterInformerFactory.Start(stopCh)
	return clusterController
}
