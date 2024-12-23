package server

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
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

func showLoginCheck(c *gin.Context, isLogin bool) {
	type DataStruct struct {
		IsLogin bool `json:"is_login"`
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "",
		"data":    DataStruct{IsLogin: isLogin},
	})
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
		if GetToken(c) != GlobalConfig.Server.Token {
			showNotAuthorized(c, "invalid token")
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetToken(c *gin.Context) string {
	// try get from header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		headerParts := strings.SplitN(authHeader, " ", 2)
		if len(headerParts) == 2 && headerParts[0] == "Bearer" {
			return headerParts[1]
		} else if len(headerParts) == 2 && headerParts[0] == "Basic" {
			userInfoBase64 := headerParts[1]
			userInfo, err := base64.StdEncoding.DecodeString(userInfoBase64)
			if err != nil {
				return ""
			}
			userInfoParts := strings.Split(string(userInfo), ":")
			if len(userInfoParts) == 2 {
				return userInfoParts[1]
			}
		}
	}
	// try get from cookie
	token, _ := c.Cookie("opstoken")
	return token
}
