package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"zeus/internal/service"
	"zeus/pkg/pagination"
	"zeus/pkg/util/httputil/response"
)

type NamespaceHandler interface {
	List(c *gin.Context)
	Get(c *gin.Context)
}

type namespaceHandler struct {
	service service.NamespaceService
}

func NewNamespaceHandler(service service.NamespaceService) *namespaceHandler {
	return &namespaceHandler{service: service}
}

// List
// @Description 获取集群命名空间列表
// @Tags namespaces
// @Accept json
// @Produce json
// @Param cluster path string true "集群名称"
// @Param labelSelector query string false "labelSelector"
// @Param page query int false "page"
// @Param limit query int false "limit"
// @Success 200 {object} response.SuccessResponse "success"
// @Router /clusters/{cluster}/namespaces [get]
func (h *namespaceHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")
	labelSelector := c.Query("labelSelector")
	pageParams, err := pagination.FromRequest(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	namespaceList, err := h.service.List(clusterName, labelSelector, pageParams)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(namespaceList))
}

// Get
// @Description 获取集群命名空间
// @Tags namespaces
// @Accept json
// @Produce json
// @Param cluster path string true "集群名称"
// @Param namespace path string true "命名空间"
// @Success 200 {object} response.SuccessResponse "success"
// @Router /clusters/{cluster}/namespaces/{namespace} [get]
func (h *namespaceHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespaceName := c.Param("namespace")
	namespace, err := h.service.Get(clusterName, namespaceName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(namespace))
}
