package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type GoogleUser struct {
	Name string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
	Profile string `json:"profile" binding:"required"`
	Given_name string `json:"firstName" binding:"required"`
	Family_name string `json:"lastName" binding:"required"`
	Email_verified bool `json:"emailVerified" binding:"required"`
}

type GoogleAuthResult struct {
	credential string
	select_by string
}

func GoogleLogin(c *gin.Context) {
	var authResult GoogleAuthResult
	err := c.BindJSON(&authResult)
	if err != nil {
		c.JSON(http.StatusBadRequest, "")
	}

	user, err := parseJwtToken(authResult.credential)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "")
	}

	jwtForUser, err := buildJwtTokenForUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "while redirecting")
	}

	appUrl := os.Getenv("APP_URL")

	c.SetCookie(HttpCookie, jwtForUser, 1*60*60, "/", appUrl, true, true) //set cookie for one hour 
	c.Redirect(200, fmt.Sprintf("%s/welcome/get-started?step=1&user=%s&obl=%s" ,appUrl, user.Name, jwtForUser))
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
		user.Email = fmt.Sprint(claims["email"])
		user.Name = fmt.Sprint(claims["name"])
		user.Profile = fmt.Sprint(claims["profile"])
		user.Given_name = fmt.Sprint(claims["given_name"])
		user.Family_name = fmt.Sprint(claims["family_name"])
		verified, _ :=  strconv.ParseBool(fmt.Sprintf("%t", claims["email_verified"]))
		user.Email_verified = verified

		return &user, nil
	} else {
		return nil, err
	}

	return nil, nil
}
