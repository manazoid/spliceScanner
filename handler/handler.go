package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func GetRouter() *gin.Engine {
	router := gin.Default()

	//CORS Policy
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "POST", "GET", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authority", "Content-Type", "Access-Control-Allow-Headers", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.SetTrustedProxies(nil)

	// handle Get method for /json
	router.GET("/:sess/:login/:pass", JsonApi)

	return router
}
