package persistence

import (
	"github.com/antonioalfa22/go-rest-template/internal/pkg/db"
	models "github.com/antonioalfa22/go-rest-template/internal/pkg/models/tokens"
	"github.com/google/uuid"
)

type TokenRepository struct{}

var tokenRepository *TokenRepository

func GetTokenRepository() *TokenRepository {
	if tokenRepository == nil {
		tokenRepository = &TokenRepository{}
	}
	return tokenRepository
}

func (r *TokenRepository) GetForUser(userID uuid.UUID, tokenType string) (*models.Token, error) {
	var token models.Token
	where := models.Token{}
	where.TokenType = tokenType
	where.UserID = userID
	_, err := First(&where, &token, []string{})
	if err != nil {
		return nil, err
	}
	return &token, err
}

func (r *TokenRepository) Add(token *models.Token) error {
	err := Save(&token)
	return err
}

func (r *TokenRepository) Update(token *models.Token, userID uuid.UUID) error {
	var existing models.Token
	_, err := First(models.Token{UserID: userID, TokenType: token.TokenType}, &token, []string{})
	if err != nil {
		return err
	}
	existing.TokenType = token.TokenType
	existing.HashedToken = token.HashedToken
	existing.HashedRefreshToken = token.HashedRefreshToken
	err = Save(&existing)
	return err
}

func (r *TokenRepository) Delete(token *models.Token) error {
	err := db.GetDB().Unscoped().Delete(&token).Error
	return err
}
