package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/navisidhu/logowl/internal/controllers"
	"github.com/navisidhu/logowl/internal/middlewares"
	"github.com/navisidhu/logowl/internal/store"
)

func userRoutes(router *gin.RouterGroup, store store.InterfaceStore) {
	router.Use(middlewares.VerifyUserJwt(store))

	controller := controllers.GetUserController(store)

	router.GET("/", controller.Get)
	router.POST("/invite", controller.Invite)
	router.DELETE("/", controller.DeleteUserAccount)
	router.DELETE("/:id", controller.Delete)
}
