package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/antonioalfa22/go-rest-template/internal/pkg/persistence"
	"github.com/antonioalfa22/go-rest-template/pkg/http-err"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @GetUserCalendars godoc
// @Summary Retrieves connected Calendars for a user
// @Description get Calendar for logged in user
// @Produce json
// @Success 200 {object} calendars.Calendar
// @Router /user/connected-calendars [get]
// @Security Http Only Cookie

func GetUserCalendars(c *gin.Context) {
	s := persistence.GetCalendarRepository()

	userID := c.GetString("LoggedInUserID")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http_err.NewError(c, http.StatusBadRequest, errors.New("no valid user"))
		c.AbortWithStatus(400)
		return
	}

	if calendars, err := s.GetForUser(userUUID); err != nil {
		http_err.NewError(c, http.StatusNotFound, errors.New("calendar not found"))
		log.Println(err)
		c.AbortWithStatus(404)
	} else {
		c.JSON(http.StatusOK, calendars)
	}
}

// GetUserCalendarByID godoc
// @Summary Retrieves calendar based on given ID
// @Description get Calendar by ID
// @Produce json
// @Param id path integer true "User ID"
// @Success 200 {object} calendars.Calendar
// @Router /user/connected-calendars/{id} [get]
// @Security Http Cookie

func GetUserCalendarByID(c *gin.Context) {
	s := persistence.GetCalendarRepository()
	id := c.Param("id")
	calID, err := uuid.Parse(id)
	if err != nil {
		http_err.NewError(c, http.StatusBadRequest, errors.New("invalid request"))
		c.AbortWithStatus(400)
		return
	}

	if calendar, err := s.Get(calID); err != nil {
		http_err.NewError(c, http.StatusNotFound, errors.New("calendar not found"))
		log.Println(err)
		c.AbortWithStatus(404)
		return
	} else {
		c.JSON(http.StatusOK, calendar)
	}
}
