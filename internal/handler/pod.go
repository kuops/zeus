package handler

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/util/flushwriter"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"syscall"
	"zeus/internal/service"
	"zeus/pkg/pagination"
	"zeus/pkg/util/httputil/response"
)

type PodHandler interface {
	List(c *gin.Context)
	Get(c *gin.Context)
	Log(c *gin.Context)
	Exec(c *gin.Context)
}

type podHandler struct {
	service service.PodService
}

func NewPodHandler(service service.PodService) *podHandler {
	return &podHandler{
		service: service,
	}
}

// List
// @Description 获取集群 pod 列表
// @Tags pod
// @Accept json
// @Produce json
// @Param cluster path string true "集群名称"
// @Param ip query string false "ip"
// @Param labelSelector query string false "labelSelector"
// @Param namespace query string false "namespace"
// @Param page query int false "page"
// @Param limit query int false "limit"
// @Success 200 {object} response.SuccessResponse "success"
// @Router /clusters/{cluster}/pods [get]
func (h *podHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Query("namespace")
	labelSelector := c.Query("labelSelector")
	pageParams, err := pagination.FromRequest(c)
	podIP := c.Query("ip")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	podList, err := h.service.List(clusterName, labelSelector, namespace, podIP, pageParams)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(podList))
}

// Get
// @Description 获取集群 pod
// @Tags pod
// @Accept json
// @Produce json
// @Param cluster path string true "集群名称"
// @Param namespace path string true "namespace"
// @Param pod path string true "pod 名称"
// @Success 200 {object} response.SuccessResponse "success"
// @Router /clusters/{cluster}/namespaces/{namespace}/pods/{pod} [get]
func (h *podHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	pod, err := h.service.Get(clusterName, namespace, podName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(pod))
}

// Log
// @Description 获取集群 pod 日志
// @Tags pod
// @Accept json
// @Produce json
// @Param cluster path string true "集群名称"
// @Param namespace path string true "namespace"
// @Param pod path string true "pod 名称"
// @Param container query string false "container 名称"
// @Param follow query bool false "是否跟踪"
// @Param tailLines query int false "tail"
// @Success 200 {string} string "success"
// @Router /clusters/{cluster}/namespaces/{namespace}/pods/{pod}/log [get]
func (h *podHandler) Log(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	container := c.Query("container")
	follow, _ := strconv.ParseBool(c.DefaultQuery("follow", "false"))
	tailLines, _ := strconv.ParseInt(c.DefaultQuery("tailLines", "50"), 10, 64)
	logOptions := &v1.PodLogOptions{
		TypeMeta:  metav1.TypeMeta{},
		Container: container,
		Follow:    follow,
		TailLines: &tailLines,
	}

	stream, err := h.service.Log(clusterName, namespace, podName, logOptions,c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	defer stream.Close()

	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.WriteHeader(http.StatusOK)
	writer := flushwriter.Wrap(c.Writer)
	_,err = io.Copy(writer,stream)
	if err != nil && !errors.Is(err,syscall.EPIPE) && !errors.Is(err,context.Canceled) {
		c.AbortWithStatusJSON(http.StatusInternalServerError,response.ErrorInternalError(err.Error()))
		return
	}
}

func (h *podHandler) Exec(c *gin.Context)  {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	container := c.Query("container")

	conn,err := wsUpGrader.Upgrade(c.Writer,c.Request,nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,response.ErrorInternalError(err.Error()))
		return
	}
	defer conn.Close()

	err = h.service.Exec(clusterName,namespace,podName,container,conn)
	if err != nil {
		klog.Error(err)
		conn.Close()
		return
	}
}

var wsUpGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

