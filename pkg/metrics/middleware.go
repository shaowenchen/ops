/*
Copyright 2022 shaowenchen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware returns a Gin middleware for collecting HTTP metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Process request
		c.Next()

		// Get status code
		statusCode := strconv.Itoa(c.Writer.Status())

		// Record metrics
		HTTPRequestsTotal.WithLabelValues(PodName, c.Request.Method, path, statusCode).Inc()
	}
}

// RecordAPIRequest records an API request
func RecordAPIRequest(endpoint, namespace, status string) {
	APIRequestsTotal.WithLabelValues(PodName, endpoint, namespace, status).Inc()
}

// RecordAPIError records an API error
func RecordAPIError(endpoint, namespace, errorType string) {
	APIErrorsTotal.WithLabelValues(PodName, endpoint, namespace, errorType).Inc()
}
