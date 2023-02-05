package users

import (
	"github.com/antonioalfa22/go-rest-template/internal/pkg/models"
	tokens "github.com/antonioalfa22/go-rest-template/internal/pkg/models/tokens"
	"time"
)

type User struct {
	models.Model
	Username      string    `gorm:"column:username;" json:"username" form:"username"`
	Firstname     string    `gorm:"column:firstname;not null;" json:"firstname" form:"firstname"`
	Lastname      string    `gorm:"column:lastname;" json:"lastname" form:"lastname"`
	PasswordHash  string    `gorm:"column:password;"`
	LoginEmail    string    `gorm:"column:login_email;not null;unique_index:email" json:"email" form:"email"`
	LoginProvider string    `gorm:"column:login_provider;not null;" json:"provider" form:"provider"`
	LastLogin     time.Time `gorm:"column:last_login;"`
	Tokens        []tokens.Token
}

func (m *User) BeforeCreate() error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *User) BeforeUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}
