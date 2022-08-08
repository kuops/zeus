package repository

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"zeus/pkg/kubernetes/clusterresource"
)

type NamespaceRepository interface {
	List(clusterName, labelSelector string) ([]*v1.Namespace, error)
	Get(clusterName, namespaceName string) (*v1.Namespace, error)
}

type namespaceRepository struct {
	clusterResources map[string]*clusterresource.ClusterResource
}

func NewNamespaceRepository(clusterResources map[string]*clusterresource.ClusterResource) *namespaceRepository {
	return &namespaceRepository{
		clusterResources: clusterResources,
	}
}

func (r *namespaceRepository) List(clusterName, labelSelector string) ([]*v1.Namespace, error) {
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		return nil, err
	}
	if _, ok := r.clusterResources[clusterName]; !ok {
		return nil, errors.New("cluster resources not exists")
	}
	return r.clusterResources[clusterName].NamespaceLister.List(selector)
}

func (r *namespaceRepository) Get(clusterName, namespaceName string) (*v1.Namespace, error) {
	if _, ok := r.clusterResources[clusterName]; !ok {
		return nil, errors.New("cluster resources not exists")
	}
	return r.clusterResources[clusterName].NamespaceLister.Get(namespaceName)
}
