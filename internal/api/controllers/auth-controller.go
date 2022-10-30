package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type LoggedInUser struct {
	user GoogleUser
}

func ValidateLoggedIn(c *gin.Context) {
	var user LoggedInUser

	jwtToken, err := c.Cookie(HttpCookie)

	if err != nil {
		c.JSON(http.StatusUnauthorized, "")
	}

	googleUser, err := parseJwtTokenForLoggedInUser(jwtToken)

	if err != nil {
		c.JSON(http.StatusInternalServerError, "")
	}

	user.user = *googleUser
	c.JSON(200, gin.H{"data": user})
}

func parseJwtTokenForLoggedInUser(tokenString string) (*GoogleUser, error) {

	secret := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var user GoogleUser

		if err := json.Unmarshal([]byte(claims["user"].(string)), &user); err != nil {
			return nil, err
		}
		
		return &user, nil

	} else {
		return nil, err
	}
}

/*
func Login(c *gin.Context) {
	var loginInput LoginInput
	_ = c.BindJSON(&loginInput)
	s := persistence.GetUserRepository()
	if user, err := s.GetByUsername(loginInput.Username); err != nil {
		http_err.NewError(c, http.StatusNotFound, errors.New("user not found"))
		log.Println(err)
	} else {
		if !crypto.ComparePasswords(user.Hash, []byte(loginInput.Password)) {
			http_err.NewError(c, http.StatusForbidden, errors.New("user and password not match"))
			return
		}
		token, _ := crypto.CreateToken(user.Username)
		c.JSON(http.StatusOK, token)
	}
}
*/
