//go:generate goversioninfo -icon=favicon.ico -manifest=main.manifest
package main

import (
	"crypto/tls"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"websocket-splice/handler"
)

func main() {
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(mode)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if err := handler.Start(); err != nil {
		handler.LogError(err.Error())
		return
	}

	router := handler.GetRouter()

	port := os.Getenv("PORT")
	if port == "" {
		router.Run(":8081")
	} else {
		if err := router.Run(":" + port); err != nil {
			handler.LogError(err.Error())
			return
		}
	}
}
