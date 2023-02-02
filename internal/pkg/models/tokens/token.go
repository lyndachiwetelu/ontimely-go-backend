package tokens

import (
	"github.com/antonioalfa22/go-rest-template/internal/pkg/models"
	"time"
)

type Token struct {
	models.Model
	UserID int  `gorm:"column:user_id;not null;" json:"user_id" form:"user_id"`
	TokenType string  `gorm:"column:token_type;not null;" json:"token_type" form:"token_type"`
	HashedToken string  `gorm:"column:hashed_token;" json:"hashed_token" form:"hashed_token"`
}

func (m *Token) BeforeCreate() error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Token) BeforeUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}
