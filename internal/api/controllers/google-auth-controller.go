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


type OntimelyClaims struct {
	user GoogleUser `json:"user"`
	jwt.RegisteredClaims
}


func GoogleLogin(c *gin.Context) {
	credential := c.PostForm("credential")

	if credential == "" {
		c.JSON(http.StatusBadRequest, "")
	}

	user, err := verifyToken(credential)

	if err != nil {
		//redirect to error page on client
		c.JSON(http.StatusForbidden, fmt.Sprintf("Invalid gtkn %v", err))
	}

	//user, err := parseJwtToken(credential)

	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("tkn server error %v", err))
		return
	}

	/*jwtForUser, err := buildJwtTokenForUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("error while redirecting tkn %v", err))
		return
	}*/

	appUrl := os.Getenv("APP_URL")

	c.SetCookie(HttpCookie, "jwtForUser", 1*60*60, "/", appUrl, true, true) //set cookie for one hour
	c.Redirect(200, fmt.Sprintf("%s/welcome/get-started?step=1&user=%s&obl=%s", appUrl, user.Name, "jwtForUser"))
}

func verifyToken(token string) (*GoogleUser, error) {
	ctx := context.Background()
	audience := os.Getenv("GOOGLE_CLIENT_ID")

	payload, err := idtoken.Validate(ctx, token, audience)

	if err != nil {
		return nil, errors.New("error occurred: invalid_tkn")
	}

	var user GoogleUser
	user.Email = payload.Claims["email"].(string)
	user.Name = payload.Claims["name"].(string)
	user.Family_name = payload.Claims["family_name"].(string)
	user.Given_name = payload.Claims["given_name"].(string)
	user.Email_verified = payload.Claims["email_verified"].(bool)

	return &user, nil
}

func buildJwtTokenForUser(user *GoogleUser) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	// Create the claims
	claims := OntimelyClaims{
		*user,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Issuer:    "https://ontimelyapp.com",
			ID:        "1",
			Audience:  []string{"ontimelyapp"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//ss, err := token.SignedString(secret)
	//fmt.Printf("%v %v", ss, err)

	/*token := jwt.New(jwt.SigningMethodEdDSA)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(1 * time.Hour)
	claims["authorized"] = true
	claims["user"] = user*/

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}