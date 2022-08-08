package service

import (
	"zeus/internal/repository"
)

type Services struct {
	Cluster   ClusterService
	Pod       PodService
	Namespace NamespaceService
	Node      NodeService
}

func NewServices(repos *repository.Repositories) *Services {
	return &Services{
		Cluster:   NewClusterService(repos.Cluster),
		Pod:       NewPodService(repos.Pod),
		Namespace: NewNamespaceService(repos.Namespace),
		Node:      NewNodeService(repos.Node),
	}
}
