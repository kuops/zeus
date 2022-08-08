package service

import (
	v1 "k8s.io/api/core/v1"
	"sort"
	"zeus/internal/repository"
	"zeus/pkg/pagination"
)

type NamespaceService interface {
	List(clusterName, labelSelector string, pageParams *pagination.PageParams) (*pagination.PageData, error)
	Get(clusterName, namespaceName string) (*v1.Namespace, error)
}

type namespaceService struct {
	repo repository.NamespaceRepository
}

func NewNamespaceService(repo repository.NamespaceRepository) *namespaceService {
	return &namespaceService{
		repo: repo,
	}
}

func (s *namespaceService) List(clusterName, labelSelector string, pageParams *pagination.PageParams) (*pagination.PageData, error) {
	var data = pagination.PageData{}
	namespaceList, err := s.repo.List(clusterName, labelSelector)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(namespaceList, func(i, j int) bool {
		return namespaceList[i].Name < namespaceList[j].Name
	})
	paginator, offset, end, err := pagination.Paginate(len(namespaceList), pageParams)
	if err != nil {
		return nil, err
	}
	data.Items = namespaceList[offset:end]
	data.Paginator = *paginator
	return &data, nil
}

func (s *namespaceService) Get(clusterName, namespaceName string) (*v1.Namespace, error) {
	return s.repo.Get(clusterName, namespaceName)
}
