package repository

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "zeus/pkg/kubernetes/apis/cluster/v1"
	clusterclientset "zeus/pkg/kubernetes/client/clientset/versioned"
	"zeus/pkg/kubernetes/clusterresource"
)

type ClusterRepository interface {
	List(name,provider string) (*clusterv1.ClusterList, error)
	Get(name string) (*clusterv1.Cluster, error)
}

type clusterRepository struct {
	clientSet        clusterclientset.Interface
	clusterResources map[string]*clusterresource.ClusterResource
}

func NewClusterRepository(clientSet clusterclientset.Interface,
	clusterResources map[string]*clusterresource.ClusterResource) *clusterRepository {
	return &clusterRepository{
		clientSet:        clientSet,
		clusterResources: clusterResources,
	}
}

func (r *clusterRepository) List(name,provider string) (*clusterv1.ClusterList, error) {
	var fieldSelector string
	if name != "" {
		fieldSelector = fmt.Sprintf("metadata.name=%s",name)
	}
	clusterList, err := r.clientSet.ClusterV1().Clusters().List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		return nil, err
	}
	if provider != "" {
		var filteredItems []clusterv1.Cluster
		for _,clusterItem := range clusterList.Items {
			if clusterItem.Status.Provider == provider {
				filteredItems = append(filteredItems, clusterItem)
			}
		}
		clusterList.Items = filteredItems
	}
	return clusterList,nil
}

func (r *clusterRepository) Get(name string) (*clusterv1.Cluster, error) {
	return r.clientSet.ClusterV1().Clusters().Get(context.TODO(), name, metav1.GetOptions{})
}
