package repository

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"strings"
	"zeus/pkg/kubernetes/clusterresource"
)

type NodeRepository interface {
	List(clusterName, labelSelector, nodeIP string) ([]*v1.Node, error)
	Get(clusterName, nodeName string) (*v1.Node, error)
}

type nodeRepository struct {
	clusterResources map[string]*clusterresource.ClusterResource
}

func NewNodeRepository(clusterResources map[string]*clusterresource.ClusterResource) *nodeRepository {
	return &nodeRepository{
		clusterResources: clusterResources,
	}
}

func (r *nodeRepository) List(clusterName, labelSelector, nodeIP string) ([]*v1.Node, error) {
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		return nil, err
	}
	if _, ok := r.clusterResources[clusterName]; !ok {
		return nil, errors.New("cluster resources not exists")
	}
	nodeList, err := r.clusterResources[clusterName].NodeLister.List(selector)

	if nodeIP != "" {
		var nodeListByIp []*v1.Node
		for _, pod := range nodeList {
			for _, address := range pod.Status.Addresses {
				if address.Type == "InternalIP" {
					if strings.Contains(address.Address, nodeIP) {
						nodeListByIp = append(nodeListByIp, pod)
					}
				}
			}
		}
		return nodeListByIp, nil
	}

	return nodeList, err
}

func (r *nodeRepository) Get(clusterName, nodeName string) (*v1.Node, error) {
	if _, ok := r.clusterResources[clusterName]; !ok {
		return nil, errors.New("cluster resources not exists")
	}
	return r.clusterResources[clusterName].NodeLister.Get(nodeName)
}
