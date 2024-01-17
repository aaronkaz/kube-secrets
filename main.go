package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Any("/", func(c *gin.Context) {
		r := make(map[string]string)
		r["foo"] = "bar"
		c.JSON(http.StatusOK, r)

	})
	r.Run(":9000")
}
