package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"zeus/internal/service"
	"zeus/pkg/pagination"
	"zeus/pkg/util/httputil/response"
)

type ClusterHandler interface {
	List(c *gin.Context)
	Get(c *gin.Context)
}

type clusterHandler struct {
	service service.ClusterService
}

func NewClusterHandler(service service.ClusterService) *clusterHandler {
	return &clusterHandler{service: service}
}

// List
// @Description 获取集群信息列表
// @Tags clusters
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse "success"
// @Router /clusters [get]
func (h *clusterHandler) List(c *gin.Context) {
	name := c.Query("name")
	provider := c.Query("provider")
	pageParams, err := pagination.FromRequest(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	clusterList, err := h.service.List(pageParams,name,provider)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(clusterList))
}

// Get
// @Description 获取集群信息
// @Tags clusters
// @Accept json
// @Produce json
// @Param cluster path string true "集群名称"
// @Success 200 {object} response.SuccessResponse "success"
// @Router /clusters/{cluster} [get]
func (h *clusterHandler) Get(c *gin.Context) {
	name := c.Param("cluster")
	cluster, err := h.service.Get(name)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorInternalError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse(cluster))
}
