package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthMessage struct {
	Message string `json:"message"`
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthMessage{Message: "api is connected!"})
}