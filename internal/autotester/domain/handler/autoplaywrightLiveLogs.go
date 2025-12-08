package handler

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleLogRequest handles a request with SSE
func (a *AutotesterController) HandleLogRequest(c *gin.Context) {
	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Writer.CloseNotify():
			return false
		default:
			c.SSEvent("hallo", "tschau")
			time.Sleep(time.Second)
			return true
		}
	})
}
