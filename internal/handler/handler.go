package handler

import "zeus/internal/service"

type Handlers struct {
	Cluster   ClusterHandler
	Pod       PodHandler
	Namespace NamespaceHandler
	Node      NodeHandler
}

func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		Cluster:   NewClusterHandler(services.Cluster),
		Pod:       NewPodHandler(services.Pod),
		Namespace: NewNamespaceHandler(services.Namespace),
		Node:      NewNodeHandler(services.Node),
	}
}
