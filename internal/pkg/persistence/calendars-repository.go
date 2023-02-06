package persistence

import (
	models "github.com/antonioalfa22/go-rest-template/internal/pkg/models/calendars"
	"github.com/google/uuid"
)

type CalendarRepository struct{}

var calendarRepository *CalendarRepository

func GetCalendarRepository() *CalendarRepository {
	if calendarRepository == nil {
		calendarRepository = &CalendarRepository{}
	}
	return calendarRepository
}

func (r *CalendarRepository) GetForUser(userID uuid.UUID) (*[]models.Calendar, error) {
	var calendars []models.Calendar
	where := models.Calendar{}
	where.Hidden = false
	where.Deleted = false
	where.UserID = userID

	err := Find(where, &calendars, []string{}, "created_at asc")
	return &calendars, err
}

func (r *CalendarRepository) GetTypeForUser(userID uuid.UUID, calendarType string) (*[]models.Calendar, error) {
	var calendars []models.Calendar
	where := models.Calendar{}
	where.CalendarType = calendarType
	where.Hidden = false
	where.Deleted = false
	where.UserID = userID

	err := Find(where, &calendars, []string{}, "created_at asc")
	return &calendars, err
}

func (r *CalendarRepository) Get(id uuid.UUID) (*models.Calendar, error) {
	var calendar models.Calendar
	where := models.Calendar{}
	where.ID = id

	_, err := First(where, &calendar, []string{})
	return &calendar, err
}

func (r *CalendarRepository) Add(calendar *models.Calendar) error {
	err := Save(&calendar)
	return err
}
