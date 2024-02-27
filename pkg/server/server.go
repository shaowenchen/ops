package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Pagination[T any] struct {
	PageSize uint `form:"page_size" json:"page_size"`
	Page     uint `form:"page" json:"page"`
	List     []T  `json:"list"`
	Total    uint `json:"total"`
}

func paginator[T any](dataList []T, pageSize, page uint) (pagination Pagination[T]) {
	pagination.PageSize = pageSize
	pagination.Page = page
	pagination.Total = uint(len(dataList))

	maxPage := pagination.Total / pageSize
	if pagination.Total%pageSize != 0 {
		maxPage++
	}
	if page > maxPage || page == 0 {
		return
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > pagination.Total {
		end = pagination.Total
	}
	if end > start {
		pagination.List = dataList[start:end]
	}
	return
}

func showNotAuthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    -1,
		"message": "not authorized, " + message,
	})
}
func showError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": message,
	})
}

func showData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func showDataSouceCopilot(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func showSuccess(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GlobalConfig.Server.Token == "" {
			c.Next()
			return
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			showNotAuthorized(c, "empty Authorization")
			c.Abort()
			return
		}

		headerParts := strings.SplitN(authHeader, " ", 2)
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			showNotAuthorized(c, "invalid Authorization")
			c.Abort()
			return
		}

		if headerParts[1] != GlobalConfig.Server.Token {
			showNotAuthorized(c, "invalid token")
			c.Abort()
			return
		}

		c.Next()
	}
}
