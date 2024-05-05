package routes

import (
	"rest/api/controllers"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine) {
	router.POST("/auth/signup", controllers.HandleSignup)
	router.POST("/auth/login", controllers.HandleLogin)
	router.GET("/documentation", controllers.GetDocumentation)
	router.GET("/parkings", controllers.GetParkings)
	router.POST("/parkings/closest", controllers.GetClosestParking)
	router.GET("/parkings/:id", controllers.GetParking)
	router.POST("/parkings/:spotId/:userId/:duration", controllers.ReserveSpot)
	router.GET("/users/:id", controllers.CheckReserved)
	router.POST("/users/:id/:amount", controllers.AddMoney)
}
