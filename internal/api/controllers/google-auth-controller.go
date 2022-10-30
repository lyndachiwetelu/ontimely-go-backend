package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/api/idtoken"
)

type GoogleUser struct {
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required"`
	Profile        string `json:"profile" binding:"required"`
	Given_name     string `json:"firstName" binding:"required"`
	Family_name    string `json:"lastName" binding:"required"`
	Email_verified bool   `json:"emailVerified" binding:"required"`
}

type GoogleAuthResult struct {
	credential string
}

func GoogleLogin(c *gin.Context) {

	var authResult GoogleAuthResult
	err := c.Bind(&authResult)
	if err != nil {
		c.JSON(http.StatusBadRequest, "")
	}

	err = verifyToken(authResult.credential)

	if err != nil {
		//redirect to error page on client
		c.JSON(http.StatusForbidden, "Invalid gtkn")
	}

	user, err := parseJwtToken(authResult.credential)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "tkn server error")
	}

	jwtForUser, err := buildJwtTokenForUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "error while redirecting tkn")
	}

	appUrl := os.Getenv("APP_URL")

	c.SetCookie(HttpCookie, jwtForUser, 1*60*60, "/", appUrl, true, true) //set cookie for one hour
	c.Redirect(200, fmt.Sprintf("%s/welcome/get-started?step=1&user=%s&obl=%s", appUrl, user.Name, jwtForUser))
}

func verifyToken(token string) error {
	ctx := context.Background()
	audience := os.Getenv("GOOGLE_CLIENT_ID")

	_, err := idtoken.Validate(ctx, token, audience)

	if err != nil {
		return errors.New("error occurred: invalid_tkn")
	}

	return nil
}

func buildJwtTokenForUser(user *GoogleUser) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	token := jwt.New(jwt.SigningMethodEdDSA)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(1 * time.Hour)
	claims["authorized"] = true
	claims["user"] = user
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func parseJwtToken(tokenString string) (*GoogleUser, error) {

	secret := []byte(os.Getenv("GOOGLE_OAUTH_SECRET"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var user GoogleUser
		user.Email = claims["email"].(string)
		user.Name = claims["name"].(string)
		user.Profile = claims["profile"].(string)
		user.Given_name = claims["given_name"].(string)
		user.Family_name = claims["family_name"].(string)
		verified, _ := claims["email_verified"].(bool)
		user.Email_verified = verified

		return &user, nil
	} else {
		return nil, err
	}
}
