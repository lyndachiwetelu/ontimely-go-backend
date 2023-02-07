package controllers

import (
	"errors"
	"fmt"
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

type updateCalendarRequest struct {
	Category   string `json:"category" binding:"required"`
	Name       string `json:"name" binding:"required"`
	CalendarID string `json:"id" binding:"required"`
}

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
		http_err.NewError(c, http.StatusNotFound, errors.New("calendars counld not be fetched"))
		log.Println(err)
		c.AbortWithStatus(500)
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
	decodedId, err := url.QueryUnescape(id)
	if err != nil {
		fmt.Printf("decoding state error %v", err)
	}
	encKey := os.Getenv("ENCRYPTION_KEY")
	idUUID := crypto.DecryptString(decodedId, encKey)

	log.Printf("idUUID %s original id encrypted %s",idUUID, id)

	calID, err := uuid.Parse(idUUID)
	if err != nil {
		log.Printf("%v", err)
		http_err.NewError(c, http.StatusBadRequest, errors.New("invalid calendar id in request"))
		c.AbortWithStatus(400)
		return
	}

	userID := c.GetString("LoggedInUserID")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("%v", err)
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

func UpdateUserCalendarDetail(c *gin.Context) {
	s := persistence.GetCalendarRepository()
	userUUID := checkLoggedIn(c)

	var reqBody updateCalendarRequest

	if err := c.BindJSON(&reqBody); err != nil {
		http_err.NewError(c, http.StatusNotFound, errors.New("invalid request"))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	calendarID, err := decryptCalendarIdFromRequest(reqBody.CalendarID)

	if err != nil {
		http_err.NewError(c, http.StatusNotFound, errors.New("invalid calendar id"))
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}

	calID := *calendarID
	uID := *userUUID

	if calendar, err := s.Get(calID); err != nil {
		http_err.NewError(c, http.StatusNotFound, errors.New("calendar not found"))
		log.Println(err)
		c.AbortWithStatus(404)
		return
	} else {
		if calendar.UserID != uID {
			//check user ID matches - move to some middleware for calendar routes!!
			log.Println("the user requesting to update this calendar is not the calendar owner. request is forbidden")
			http_err.NewError(c, http.StatusForbidden, errors.New("forbidden request"))
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		//udate
		calendar.Name = reqBody.Name
		calendar.Category = reqBody.Category

		if err := s.Add(calendar); err != nil {
			log.Printf("error occurred %v", err)
			http_err.NewError(c, http.StatusForbidden, errors.New("server error"))
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.JSON(http.StatusOK, calendar)
	}
}

func decryptCalendarIdFromRequest(id string) (*uuid.UUID, error) {
	decodedId, err := url.QueryUnescape(id)
	if err != nil {
		fmt.Printf("decoding state error %v", err)
		return nil, err
	}
	encKey := os.Getenv("ENCRYPTION_KEY")
	idUUID := crypto.DecryptString(decodedId, encKey)

	calID, err := uuid.Parse(idUUID)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}

	return &calID, nil
}

func checkLoggedIn(c *gin.Context) *uuid.UUID {
	userID := c.GetString("LoggedInUserID")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http_err.NewError(c, http.StatusBadRequest, errors.New("no valid user"))
		c.AbortWithStatus(400)
		return nil
	}

	return &userUUID
}
