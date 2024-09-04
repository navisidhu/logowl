package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/navisidhu/logowl/internal/controllers"
	"github.com/navisidhu/logowl/internal/store"
)

func loggingRoutes(router *gin.RouterGroup, store store.InterfaceStore) {
	controller := controllers.GetLoggingController(store)

	router.POST("/error", controller.RegisterError)
	router.POST("/analytics", controller.RegisterAnalyticEvent)
}
