package main

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed docs/openapi.yaml
var openAPISpec []byte

//go:embed docs/redoc.html
var reDocPage []byte

func registerDocumentationRoutes(router gin.IRoutes) {
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openAPISpec)
	})
	router.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", reDocPage)
	})
}
