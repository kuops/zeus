package repository

import (
	clusterclientset "zeus/pkg/kubernetes/client/clientset/versioned"
	"zeus/pkg/kubernetes/clusterresource"
)

type Repositories struct {
	Cluster   ClusterRepository
	Pod       PodRepository
	Namespace NamespaceRepository
	Node      NodeRepository
}

func NewRepositories(clientset clusterclientset.Interface,
	clusterResources map[string]*clusterresource.ClusterResource) *Repositories {
	return &Repositories{
		Cluster:   NewClusterRepository(clientset, clusterResources),
		Pod:       NewPodRepository(clusterResources),
		Namespace: NewNamespaceRepository(clusterResources),
		Node:      NewNodeRepository(clusterResources),
	}
}
