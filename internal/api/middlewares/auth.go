package middlewares

import (
	"errors"
	"log"

	"github.com/antonioalfa22/go-rest-template/internal/api/controllers"
	"github.com/antonioalfa22/go-rest-template/internal/pkg/persistence"
	http_err "github.com/antonioalfa22/go-rest-template/pkg/http-err"
	"github.com/gin-gonic/gin"
)

// func AuthRequired() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authorizationHeader := c.GetHeader("authorization")
// 		if !crypto.ValidateToken(authorizationHeader) {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 			return
// 		} else {
// 			c.Next()
// 		}
// 	}
// }

func LoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, loggedInUser, err := controllers.CheckUserThatIsLoggedIn(c)

		if err != nil {
			http_err.NewError(c, 401, errors.New("unauthorized"))
			c.AbortWithStatus(401)
			return
		}

		if loggedInUser == nil {
			http_err.NewError(c, 401, errors.New("unauthorized"))
			c.AbortWithStatus(401)
			return
		}

		// Get userID
		s := persistence.GetUserRepository()
		userInDB, err := s.GetByEmail(loggedInUser.User.Email)
		if err != nil {
			log.Printf("ALERT! could not find a logged in user in the database! %v", err)
			http_err.NewError(c, 401, errors.New("unauthorized"))
			c.AbortWithStatus(401)
			return
		}

		c.Set("LoggedInUserID", userInDB.ID.String())
		c.Next()
	}
}
