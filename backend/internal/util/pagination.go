package util

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// PageQuery 分页查询参数。
type PageQuery struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// ParsePageQuery 从 query 解析统一分页参数 page / page_size。
func ParsePageQuery(c *gin.Context) PageQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = DefaultPage
	}
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if size < 1 {
		size = DefaultPageSize
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}
	return PageQuery{Page: page, PageSize: size}
}

// PageResult 分页返回结构。
type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Paginate 为 GORM 查询附加分页（被所有 repository 复用）。
func Paginate(query *gorm.DB, pq PageQuery) *gorm.DB {
	return query.Offset((pq.Page - 1) * pq.PageSize).Limit(pq.PageSize)
}
