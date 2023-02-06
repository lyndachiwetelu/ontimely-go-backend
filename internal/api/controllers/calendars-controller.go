package controllers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/antonioalfa22/go-rest-template/internal/pkg/persistence"
	"github.com/antonioalfa22/go-rest-template/pkg/crypto"
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
	id := c.Query("id")
	idStr, _ := url.QueryUnescape(id)
	encKey := os.Getenv("ENCRYPTION_KEY")
	idUUID := crypto.DecryptString(idStr, encKey)

	calID, err := uuid.Parse(idUUID)
	if err != nil {
		http_err.NewError(c, http.StatusBadRequest, errors.New("invalid calendar id in request"))
		c.AbortWithStatus(400)
		return
	}

	userID := c.GetString("LoggedInUserID")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http_err.NewError(c, http.StatusBadRequest, errors.New("no valid user"))
		c.AbortWithStatus(400)
		return
	}

	if calendar, err := s.Get(calID); err != nil {
		http_err.NewError(c, http.StatusNotFound, errors.New("calendar not found"))
		log.Println(err)
		c.AbortWithStatus(404)
		return
	} else {

		if calendar.UserID != userUUID {
			//check user ID matches
			log.Println("the user requesting this calendar is not the calendar owner. request is forbidden")
			http_err.NewError(c, http.StatusForbidden, errors.New("forbidden request"))
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.JSON(http.StatusOK, calendar)
	}
}
