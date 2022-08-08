package service

import (
	"context"
	"github.com/gorilla/websocket"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sort"
	"zeus/internal/repository"
	"zeus/pkg/kubernetes/wsclient"
	"zeus/pkg/pagination"
)

type PodService interface {
	List(clusterName, labelSelector, namespace, podIP string, pageParams *pagination.PageParams) (*pagination.PageData, error)
	Get(clusterName, namespace, podName string) (*v1.Pod, error)
	Log(clusterName, namespace, podName string, logOptions *v1.PodLogOptions,ctx context.Context) (io.ReadCloser, error)
	Exec(clusterName,namespace, podName,container string,conn *websocket.Conn) error
}

type podService struct {
	repo repository.PodRepository
}

func NewPodService(repo repository.PodRepository) *podService {
	return &podService{
		repo: repo,
	}
}

func (s *podService) List(clusterName, labelSelector, namespace, podIP string, pageParams *pagination.PageParams) (*pagination.PageData, error) {
	var data = pagination.PageData{}
	podList, err := s.repo.List(clusterName, labelSelector, namespace, podIP)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(podList, func(i, j int) bool {
		return podList[i].Name < podList[j].Name
	})

	paginator, offset, end, err := pagination.Paginate(len(podList), pageParams)
	if err != nil {
		return nil, err
	}
	data.Items = podList[offset:end]
	data.Paginator = *paginator
	return &data, nil
}

func (s *podService) Get(clusterName, namespace, podName string) (*v1.Pod, error) {
	return s.repo.Get(clusterName, namespace, podName)
}

func (s *podService) Log(clusterName, namespace, podName string, logOptions *v1.PodLogOptions,ctx context.Context) (io.ReadCloser, error) {
	clientset, err := s.repo.GetClientSet(clusterName)
	if err != nil {
		return nil, err
	}
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	return req.Stream(ctx)
}

func (s *podService) Exec(clusterName,namespace, podName,container string,conn *websocket.Conn) error {
	clientset,err := s.repo.GetClientSet(clusterName)
	if err != nil {
		return  err
	}

	restConfig, err := s.repo.GetRestConfig(clusterName)
	if err != nil {
		return  err
	}

	wsClient := wsclient.NewWSClient(conn)
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   []string{"sh","-c","[ -x /bin/bash ] && exec /bin/bash || exec /bin/sh"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             wsClient,
		Stdout:            wsClient,
		Stderr:            wsClient,
		TerminalSizeQueue: wsClient,
		Tty:               true,
	})
	if err != nil {
		return err
	}

	return nil
}