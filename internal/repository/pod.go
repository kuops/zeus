package repository

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
	"zeus/pkg/kubernetes/clusterresource"
)

type PodRepository interface {
	List(clusterName, labelSelector, namespace, podIP string) ([]*v1.Pod, error)
	Get(clusterName, namespace, podName string) (*v1.Pod, error)
	GetClientSet(clusterName string) (kubernetes.Interface, error)
	GetRestConfig(clusterName string) (*rest.Config, error)
}

type podRepository struct {
	clusterResources map[string]*clusterresource.ClusterResource
}

func NewPodRepository(clusterResources map[string]*clusterresource.ClusterResource) *podRepository {
	return &podRepository{
		clusterResources: clusterResources,
	}
}

func (r *podRepository) List(clusterName, labelSelector, namespace, podIP string) ([]*v1.Pod, error) {
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		return nil, err
	}
	if _, ok := r.clusterResources[clusterName]; !ok {
		return nil, errors.New("cluster resources not exists")
	}
	podList, err := r.clusterResources[clusterName].PodLister.List(selector)

	if podIP != "" {
		var podListByIp []*v1.Pod
		for _, pod := range podList {
			if strings.Contains(pod.Status.PodIP, podIP) {
				podListByIp = append(podListByIp, pod)
			}
		}
		return podListByIp, nil
	}

	if namespace != "" {
		podList, err = r.clusterResources[clusterName].PodLister.Pods(namespace).List(selector)
	}
	return podList, err
}

func (r *podRepository) Get(clusterName, namespace, podName string) (*v1.Pod, error) {
	if _, ok := r.clusterResources[clusterName]; !ok {
		return nil, errors.New("cluster resources not exists")
	}
	return r.clusterResources[clusterName].PodLister.Pods(namespace).Get(podName)
}

func (r *podRepository) GetClientSet(clusterName string) (kubernetes.Interface, error) {
	if _, ok := r.clusterResources[clusterName]; !ok {
		return nil, errors.New("cluster resources not exists")
	}
	return r.clusterResources[clusterName].ClientSet, nil
}

func (r *podRepository) GetRestConfig(clusterName string) (*rest.Config, error) {
	if _, ok := r.clusterResources[clusterName]; !ok {
		return nil, errors.New("cluster resources not exists")
	}
	return r.clusterResources[clusterName].RestConfig, nil
}