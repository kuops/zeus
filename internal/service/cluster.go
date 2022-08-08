package service

import (
	"sort"
	"zeus/internal/repository"
	clusterv1 "zeus/pkg/kubernetes/apis/cluster/v1"
	"zeus/pkg/pagination"
)

type ClusterService interface {
	List(pageParams *pagination.PageParams,name,provider string) (*pagination.PageData, error)
	Get(name string) (*clusterv1.Cluster, error)
}

type clusterService struct {
	repo repository.ClusterRepository
}

func NewClusterService(repo repository.ClusterRepository) *clusterService {
	return &clusterService{
		repo: repo,
	}
}

func (s *clusterService) List(pageParams *pagination.PageParams,name,provider string) (*pagination.PageData, error) {
	var data = pagination.PageData{}
	clusterList,err := s.repo.List(name,provider)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(clusterList.Items, func(i, j int) bool {
		return clusterList.Items[i].Name < clusterList.Items[j].Name
	})

	paginator, offset, end, err := pagination.Paginate(len(clusterList.Items), pageParams)
	if err != nil {
		return nil, err
	}
	data.Items = clusterList.Items[offset:end]
	data.Paginator = *paginator
	return &data, nil
}

func (s *clusterService) Get(name string) (*clusterv1.Cluster, error) {
	return s.repo.Get(name)
}
