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
	"time"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware returns a Gin middleware for collecting HTTP metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record request size
		if c.Request.ContentLength > 0 {
			HTTPRequestSize.WithLabelValues(c.Request.Method, path).Observe(float64(c.Request.ContentLength))
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		statusCode := strconv.Itoa(c.Writer.Status())

		// Record metrics
		HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, statusCode).Inc()
		HTTPRequestDuration.WithLabelValues(c.Request.Method, path, statusCode).Observe(duration.Seconds())

		// Record response size
		if c.Writer.Size() > 0 {
			HTTPResponseSize.WithLabelValues(c.Request.Method, path, statusCode).Observe(float64(c.Writer.Size()))
		}
	}
}

// RecordAPIRequest records an API request
func RecordAPIRequest(endpoint, namespace, status string, duration time.Duration) {
	APIRequestsTotal.WithLabelValues(endpoint, namespace, status).Inc()
	APIRequestDuration.WithLabelValues(endpoint, namespace).Observe(duration.Seconds())
}

// RecordAPIError records an API error
func RecordAPIError(endpoint, namespace, errorType string) {
	APIErrorsTotal.WithLabelValues(endpoint, namespace, errorType).Inc()
}
