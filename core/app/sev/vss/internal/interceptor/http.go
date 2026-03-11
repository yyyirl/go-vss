package interceptor

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func HttpHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})

		c.Next()
	}
}

func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		finish := make(chan struct{}, 1)
		go func() {
			c.Next()
			finish <- struct{}{}
		}()

		select {
		case <-time.After(timeout):
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"error": "request timeout",
			})
		case <-finish:
		}
	}
}
