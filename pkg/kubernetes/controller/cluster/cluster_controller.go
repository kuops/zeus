package cluster

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"strings"
	"sync"
	"time"
	clusterv1 "zeus/pkg/kubernetes/apis/cluster/v1"
	clusterclientset "zeus/pkg/kubernetes/client/clientset/versioned"
	clusterinformers "zeus/pkg/kubernetes/client/informers/externalversions/cluster/v1"
	clusterlisters "zeus/pkg/kubernetes/client/listers/cluster/v1"
	"zeus/pkg/kubernetes/clusterresource"
)

type Controller struct {
	kubeClientSet    kubernetes.Interface
	clusterClientSet clusterclientset.Interface
	clusterInformer  clusterinformers.ClusterInformer
	clustersLister   clusterlisters.ClusterLister
	clustersSynced   cache.InformerSynced
	queue            workqueue.RateLimitingInterface
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder
	lock             sync.RWMutex
	clusterResources map[string]*clusterresource.ClusterResource
}

func NewController(kubeClientSet kubernetes.Interface,
	clusterClientSet clusterclientset.Interface,
	clusterInformer clusterinformers.ClusterInformer,
	clusterResources map[string]*clusterresource.ClusterResource) *Controller {
	// recorder
	broadcaster := record.NewBroadcaster()
	broadcaster.StartStructuredLogging(0)
	broadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cluster-controller"})

	c := &Controller{
		kubeClientSet:    kubeClientSet,
		clusterClientSet: clusterClientSet,
		clusterInformer:  clusterInformer,
		clustersLister:   clusterInformer.Lister(),
		clustersSynced:   clusterInformer.Informer().HasSynced,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster"),
		eventBroadcaster: broadcaster,
		eventRecorder:    recorder,
		clusterResources: clusterResources,
	}

	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueCluster,
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.enqueueCluster(newObj)
		},
		DeleteFunc: c.enqueueCluster,
	})

	return c
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.InfoS("Starting controller", "controller", "cluster")
	defer klog.InfoS("Shutting down controller", "controller", "cluster")

	if !cache.WaitForNamedCacheSync("cluster", stopCh, c.clustersSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
	return nil
}

func (c *Controller) enqueueCluster(obj interface{}) {
	cluster := obj.(*clusterv1.Cluster)
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", cluster, err))
		return
	}
	c.queue.Add(key)
}

func (c *Controller) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncCluster(key.(string))
	c.handleErr(err, key)
	return true
}

func (c *Controller) syncCluster(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("not a valid controller key %s, %#v", key, err)
		return err
	}

	cluster, err := c.clustersLister.Get(name)

	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("cluster '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	c.lock.Lock()
	clusterRes, ok := c.clusterResources[cluster.Name]
	if !ok || clusterRes == nil {
		clusterRes, err = clusterresource.BuildClusterResource(cluster.Spec.KubeConfig)
		if err != nil {
			c.lock.Unlock()
			return err
		}
		c.clusterResources[cluster.Name] = clusterRes
	}
	if !equality.Semantic.DeepEqual(clusterRes.KubeConfig, cluster.Spec.KubeConfig) {
		close(c.clusterResources[cluster.Name].StopCh)
		delete(c.clusterResources, cluster.Name)
		clusterRes, err = clusterresource.BuildClusterResource(cluster.Spec.KubeConfig)
		if err != nil {
			c.lock.Unlock()
			return err
		}
		c.clusterResources[cluster.Name] = clusterRes
	}
	c.lock.Unlock()

	err = clusterRes.ClientSet.Discovery().RESTClient().Get().AbsPath("/healthz").Do(context.TODO()).Error()
	var con clusterv1.ClusterCondition
	if err == nil {
		con = clusterv1.ClusterCondition{
			Type:               clusterv1.ClusterReady,
			Status:             v1.ConditionTrue,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             string(clusterv1.ClusterReady),
			Message:            "Cluster is available now",
		}
	} else {
		klog.Errorf("Failed connect cluster, %#v", err)
		c.eventRecorder.Eventf(cluster, v1.EventTypeWarning, "Warning", "Failed connect cluster.")
		con = clusterv1.ClusterCondition{
			Type:               clusterv1.ClusterNotReady,
			Status:             v1.ConditionFalse,
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "failed to connect get kubernetes version",
			Message:            "Cluster is not available now",
		}
	}
	c.updateClusterCondition(cluster, con)

	cluster.Status.KubernetesVersion = c.clusterResources[cluster.Name].ServerVersion
	cluster.Status.Provider = setProvider(c.clusterResources[cluster.Name].ServerVersion)
	nodes, err := c.clusterResources[cluster.Name].NodeLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to get cluster nodes, %#v", err)
		return err
	}
	cluster.Status.NodeCount = len(nodes)
	_, err = c.clusterClientSet.ClusterV1().Clusters().UpdateStatus(context.TODO(), cluster, metav1.UpdateOptions{})

	return err
}

func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 15 {
		klog.V(2).Infof("Error syncing cluster %s, retrying, %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	klog.V(4).Infof("Dropping cluster %s out of the queue.", key)
	c.queue.Forget(key)
	utilruntime.HandleError(err)
}

func (c *Controller) updateClusterCondition(cluster *clusterv1.Cluster, condition clusterv1.ClusterCondition) {
	if cluster.Status.Conditions == nil {
		cluster.Status.Conditions = make([]clusterv1.ClusterCondition, 0)
	}

	newConditions := make([]clusterv1.ClusterCondition, 0)
	for _, cond := range cluster.Status.Conditions {
		if cond.Type == condition.Type {
			continue
		}
		newConditions = append(newConditions, cond)
	}

	newConditions = append(newConditions, condition)
	cluster.Status.Conditions = newConditions
}

func (c *Controller) GetClusterClientSet() clusterclientset.Interface {
	return c.clusterClientSet
}

func setProvider(version string) string {
	switch {
	case strings.Contains(version, "tke"):
		return "Tencent"
	case strings.Contains(version, "eks"):
		return "AWS"
	default:
		return ""
	}
}
