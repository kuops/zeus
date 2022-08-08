package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"zeus/internal/service"
	"zeus/pkg/pagination"
	"zeus/pkg/util/httputil/response"
)

type NodeHandler interface {
	List(c *gin.Context)
	Get(c *gin.Context)
}

type nodeHandler struct {
	service service.NodeService
}

func NewNodeHandler(service service.NodeService) *nodeHandler {
	return &nodeHandler{service: service}
}

// List
// @Description 获取集群节点列表
// @Tags node
// @Accept json
// @Produce json
// @Param cluster path string true "集群名称"
// @Param ip query string false "ip"
// @Param labelSelector query string false "labelSelector"
// @Param page query int false "page"
// @Param limit query int false "limit"
// @Success 200 {object} response.SuccessResponse "success"
// @Router /clusters/{cluster}/nodes [get]
func (h *nodeHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")
	labelSelector := c.Query("labelSelector")
	nodeIP := c.Query("ip")
	pageParams, err := pagination.FromRequest(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	nodeList, err := h.service.List(clusterName, labelSelector, nodeIP, pageParams)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(nodeList))
}

// Get
// @Description 获取集群节点
// @Tags node
// @Accept json
// @Produce json
// @Param cluster path string true "集群名称"
// @Param node path string true "节点名称"
// @Success 200 {object} response.SuccessResponse "success"
// @Router /clusters/{cluster}/nodes/{node} [get]
func (h *nodeHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("node")
	node, err := h.service.Get(clusterName, nodeName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(node))
}
