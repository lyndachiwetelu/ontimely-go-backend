package calendars

import (
	"time"

	"github.com/antonioalfa22/go-rest-template/internal/pkg/models"
	"github.com/google/uuid"
)

type Calendar struct {
	models.Model
	Name            string    `gorm:"column:name;not null;" json:"name" form:"name"`
	Email           string    `gorm:"column:email;" json:"email" form:"email"`
	AccessRole      string    `gorm:"column:access_role;" json:"accessRole" form:"accessroles"`
	BackgroundColor string    `gorm:"column:background_color;" json:"backgroundColor" form:"backgroundColor"`
	Deleted         bool      `gorm:"column:deleted;" json:"deleted"`
	Description     string    `gorm:"column:description;" json:"description"`
	Etag            string    `gorm:"column:tag;" json:"etag"`
	ForegroundColor string    `gorm:"column:foreground_color;" json:"foregroundColor"`
	Hidden          bool      `gorm:"column:hidden;" json:"hidden"`
	CalendarId      string    `gorm:"column:calendar_id;not null;" json:"calendarId,omitempty"`
	Kind            string    `gorm:"column:kind;" json:"kind"`
	Location        string    `gorm:"column:location;" json:"location"`
	Primary         bool      `gorm:"column:is_primary;" json:"primary"`
	Summary         string    `gorm:"column:summary;" json:"summary"`
	TimeZone        string    `gorm:"column:timezone;not null;" json:"timeZone"`
	CalendarType    string    `gorm:"column:calendar_type;not null;" json:"type"`
	Category        string    `gorm:"column:category;" json:"category"`
	TokenID         uuid.UUID `gorm:"column:token_id;not null;" json:"-"`
	UserID          uuid.UUID `gorm:"column:user_id;not null;" json:"-"`
}

func (m *Calendar) BeforeCreate() error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Calendar) BeforeUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}
