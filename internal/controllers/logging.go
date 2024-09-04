package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/navisidhu/logowl/internal/models"
	"github.com/navisidhu/logowl/internal/services"
	"github.com/navisidhu/logowl/internal/store"
	"github.com/navisidhu/logowl/internal/utils"
)

type LoggingControllers struct {
	LoggingService services.InterfaceLogging
}

func (l *LoggingControllers) RegisterError(c *gin.Context) {
	errorEvent := models.Error{
		Badges:    map[string]string{},
		UserAgent: c.Request.UserAgent(),
		Count:     1,
		Timestamp: time.Now().Unix(),
	}

	err := json.NewDecoder(c.Request.Body).Decode(&errorEvent)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if !errorEvent.AnonymizeData {
		errorEvent.ClientIP = c.ClientIP()
	}

	if !errorEvent.IsValid() {
		utils.RespondWithError(c, http.StatusBadRequest, "the provided data is too large")
		return
	}

	go l.LoggingService.SaveError(errorEvent)

	utils.RespondWithSuccess(c)
}

func (l *LoggingControllers) RegisterAnalyticEvent(c *gin.Context) {
	var analyticEvent models.AnalyticEvent

	err := json.NewDecoder(c.Request.Body).Decode(&analyticEvent)
	if err != nil {
		fmt.Println(err.Error())
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if analyticEvent.Ticket == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "the ticket was not provided")
		return
	}

	analyticEvent.UserAgent = c.Request.UserAgent()

	go l.LoggingService.SaveAnalyticEvent(analyticEvent)

	utils.RespondWithSuccess(c)
}

func GetLoggingController(store store.InterfaceStore) LoggingControllers {
	loggingService := services.GetLoggingService(store)

	return LoggingControllers{
		LoggingService: &loggingService,
	}
}
