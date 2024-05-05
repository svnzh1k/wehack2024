package main

import (
	"rest/api/controllers"
	"rest/api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "DELETE", "PATCH"}
	router := gin.Default()
	router.Use(cors.New(config))
	routes.Setup(router)
	controllers.Init()
	router.Run(":8080")
}
