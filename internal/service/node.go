package service

import (
	v1 "k8s.io/api/core/v1"
	"sort"
	"zeus/internal/repository"
	"zeus/pkg/pagination"
)

type NodeService interface {
	List(clusterName, labelSelector, nodeIP string, pageParams *pagination.PageParams) (*pagination.PageData, error)
	Get(clusterName, nodeName string) (*v1.Node, error)
}

type nodeService struct {
	repo repository.NodeRepository
}

func NewNodeService(repo repository.NodeRepository) *nodeService {
	return &nodeService{
		repo: repo,
	}
}

func (s *nodeService) List(clusterName, labelSelector, nodeIP string, pageParams *pagination.PageParams) (*pagination.PageData, error) {
	var data = pagination.PageData{}
	nodeList, err := s.repo.List(clusterName, labelSelector, nodeIP)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(nodeList, func(i, j int) bool {
		return nodeList[i].Name < nodeList[j].Name
	})
	paginator, offset, end, err := pagination.Paginate(len(nodeList), pageParams)
	if err != nil {
		return nil, err
	}
	data.Items = nodeList[offset:end]
	data.Paginator = *paginator
	return &data, nil
}

func (s *nodeService) Get(clusterName, nodeName string) (*v1.Node, error) {
	return s.repo.Get(clusterName, nodeName)
}
