package pagination

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

const (
	DefaultLimit  = 10
	DefaultPage   = 1
	QueryKeyPage  = "page"
	QueryKeyLimit = "limit"
)

type PageParams struct {
	Page  int
	Limit int
}

type Paginator struct {
	Total     int `json:"total"`
	PageTotal int `json:"pageTotal"`
	Limit     int `json:"limit"`
	Page      int `json:"page"`
}

type PageData struct {
	Items interface{} `json:"items"`
	Paginator
}

func FromRequest(c *gin.Context) (*PageParams, error) {
	var pageParams PageParams
	if v := c.Query(QueryKeyPage); v != "" {
		var err error
		pageParams.Page, err = strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
	} else {
		pageParams.Page = DefaultPage
	}

	if v := c.Query(QueryKeyLimit); v != "" {
		var err error
		pageParams.Limit, err = strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		if pageParams.Limit == 0 {
			pageParams.Limit = DefaultPage
		}
	} else {
		pageParams.Limit = DefaultLimit
	}

	return &pageParams, nil
}

func Paginate(total int, pageParams *PageParams) (*Paginator, int, int, error) {
	var paginator Paginator
	if total%pageParams.Limit != 0 {
		paginator.PageTotal = total/pageParams.Limit + 1
	} else {
		paginator.PageTotal = total / pageParams.Limit
	}

	if pageParams.Page > paginator.PageTotal && paginator.PageTotal != 0 {
		paginator.Page = paginator.PageTotal
	} else {
		paginator.Page = pageParams.Page
	}

	paginator.Total = total
	paginator.Limit = pageParams.Limit
	offset := (paginator.Page - 1) * paginator.Limit
	end := offset + paginator.Limit
	if offset+paginator.Limit > total {
		end = total
	}
	return &paginator, offset, end, nil
}
