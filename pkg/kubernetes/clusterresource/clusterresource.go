package clusterresource

import (
	"fmt"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"time"
)

type ClusterResource struct {
	KubeConfig        string
	ServerVersion     string
	RestConfig        *rest.Config
	ClientSet         kubernetes.Interface
	Informer          kubeinformers.SharedInformerFactory
	NodeLister        corev1.NodeLister
	PodLister         corev1.PodLister
	ServiceLister     corev1.ServiceLister
	ConfigMapLister   corev1.ConfigMapLister
	NamespaceLister   corev1.NamespaceLister
	StatefulSetLister appsv1.StatefulSetLister
	DeploymentLister  appsv1.DeploymentLister
	DaemonSetLister   appsv1.DaemonSetLister
	StopCh            chan struct{}
}

func NewClusterResources() map[string]*ClusterResource {
	return map[string]*ClusterResource{}
}

func BuildClusterResource(kubeConfig string) (*ClusterResource, error) {
	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	if err != nil {
		klog.Errorf("Unable to create client config from kubeconfig bytes, %#v", err)
		return nil, err
	}
	clusterConfig, err := clientConfig.ClientConfig()
	if err != nil {
		klog.Errorf("Failed to get client config, %#v", err)
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		klog.Errorf("Failed to create ClientSet from config, %#v", err)
		return nil, err
	}
	serverInfo, err := clientSet.ServerVersion()
	if err != nil {
		klog.Errorf("Failed get serverInfo, %#v", err)
		return nil, err
	}

	stopCh := make(chan struct{})
	var clusterRes = &ClusterResource{
		KubeConfig:    kubeConfig,
		RestConfig:    clusterConfig,
		ServerVersion: serverInfo.GitVersion,
		ClientSet:     clientSet,
		StopCh:        stopCh,
	}

	informerFactory := kubeinformers.NewSharedInformerFactory(clientSet, time.Minute)
	nodeInformer := informerFactory.Core().V1().Nodes().Informer()
	podInformer := informerFactory.Core().V1().Pods().Informer()
	serviceInformer := informerFactory.Core().V1().Services().Informer()
	configMapInformer := informerFactory.Core().V1().ConfigMaps().Informer()
	namespaceInformer := informerFactory.Core().V1().Namespaces().Informer()
	statefulSetInformer := informerFactory.Apps().V1().StatefulSets().Informer()
	deploymentInformer := informerFactory.Apps().V1().Deployments().Informer()
	daemonSetInformer := informerFactory.Apps().V1().DaemonSets().Informer()
	informerFactory.Start(clusterRes.StopCh)

	if !cache.WaitForCacheSync(clusterRes.StopCh, nodeInformer.HasSynced, podInformer.HasSynced,
		serviceInformer.HasSynced, configMapInformer.HasSynced, namespaceInformer.HasSynced,
		statefulSetInformer.HasSynced, deploymentInformer.HasSynced, daemonSetInformer.HasSynced) {
		return nil, fmt.Errorf("failed to wait for caches to sync")
	}

	clusterRes.NodeLister = informerFactory.Core().V1().Nodes().Lister()
	clusterRes.PodLister = informerFactory.Core().V1().Pods().Lister()
	clusterRes.ServiceLister = informerFactory.Core().V1().Services().Lister()
	clusterRes.ConfigMapLister = informerFactory.Core().V1().ConfigMaps().Lister()
	clusterRes.NamespaceLister = informerFactory.Core().V1().Namespaces().Lister()
	clusterRes.StatefulSetLister = informerFactory.Apps().V1().StatefulSets().Lister()
	clusterRes.DeploymentLister = informerFactory.Apps().V1().Deployments().Lister()
	clusterRes.DaemonSetLister = informerFactory.Apps().V1().DaemonSets().Lister()

	if err != nil {
		klog.Errorf("", err)
	}

	return clusterRes, nil
}
