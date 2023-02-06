package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type LoggedInUser struct {
	User OntimelyUser `json:"user" binding:"required"`
}

func CheckUserThatIsLoggedIn(c *gin.Context) (int, *LoggedInUser, error) {
	var loogedInUser LoggedInUser

	jwtToken, err := c.Cookie(HttpCookie)

	if err != nil {
		return http.StatusUnauthorized, nil, nil
	}

	googleUser, err := parseJwtTokenForLoggedInUser(jwtToken)

	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	loogedInUser.User = *googleUser

	return http.StatusOK, &loogedInUser, nil
}

func ValidateLoggedIn(c *gin.Context) {
	status, loogedInUser, err := CheckUserThatIsLoggedIn(c)

	if err != nil {
		c.JSON(status, "")
	}

	if loogedInUser == nil {
		log.Println("could not retrieve logged in user")
		c.JSON(500, "")
		return
	}

	c.JSON(200, gin.H{"data": *loogedInUser})
}

func parseJwtTokenForLoggedInUser(tokenString string) (*OntimelyUser, error) {

	secret := []byte(os.Getenv("JWT_SECRET"))

	claims := &OntimelyClaims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	return &claims.User, nil

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
