package controllers

import (
	"context"

	// "encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/antonioalfa22/go-rest-template/internal/pkg/models/tokens"
	"github.com/antonioalfa22/go-rest-template/internal/pkg/persistence"
	"github.com/antonioalfa22/go-rest-template/pkg/crypto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type GoogleCalendar struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

func ConnectGoogleCalendar(ctx *gin.Context) {
	appUrl := os.Getenv("APP_URL")
	// check login
	status, loggedInUser, err := CheckUserThatIsLoggedIn(ctx)
	if err != nil || status != http.StatusOK {
		log.Printf("unable to read logged in user: %v", err)
		ctx.Redirect(302, appUrl)
	}

	u := persistence.GetUserRepository()
	user, err := u.GetByEmail(loggedInUser.User.Email)

	if err != nil {
		log.Printf("unable to read logged in user details from db: %v", err)
		ctx.Redirect(302, appUrl)
	}

	RequestPermission(ctx, user.ID)
}

func getClientForUser(config *oauth2.Config, tok *oauth2.Token) *http.Client {
	// get the saved token for the user
	return config.Client(context.Background(), tok)
}

func RequestPermission(ctx *gin.Context, userID uuid.UUID) {
	b, err := os.ReadFile("calendar-credentials.json")
	if err != nil {
		//remove fatals
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		//remove fatals
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	userIDString := crypto.EncryptString(userID.String(), os.Getenv("ENCRYPTION_KEY"))
	authURL := config.AuthCodeURL(userIDString, oauth2.AccessTypeOffline)
	ctx.Redirect(302, authURL)
}

func HandleGoogleAuthorizeCalendar(ctx *gin.Context) {
	appUrl := os.Getenv("APP_URL")
	token, err := handleGoogleAuthorize(ctx)

	if err != nil {
		log.Printf("unable to get user token %v", err)
		ctx.Redirect(302, appUrl)
	}

	GetCalendarInformation(ctx, token)
	ctx.Redirect(302, appUrl+"/user/dashboard/calendar/connected")
}

func handleGoogleAuthorize(ctx *gin.Context) (*oauth2.Token, error) {
	b, _ := os.ReadFile("calendar-credentials.json")
	code := ctx.Query("code")
	state := ctx.Query("state")

	if code == "" {
		err := errors.New("no code")
		return nil, err
	}

	if state == "" {
		err := errors.New("no state parameter. invalid request")
		return nil, err
	}

	config, _ := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Printf("Unable to retrieve token from web: %v", err)
		return nil, err
	}

	u := persistence.GetUserRepository()
	userIDStr := crypto.DecryptString(state, os.Getenv("ENCRYPTION_KEY"))
	userId, err := uuid.Parse(userIDStr)

	if err != nil {
		log.Printf("uuid parsing error: %v", err)
		return nil, err
	}

	user, err := u.Get(userId)
	if err != nil {
		return nil, err
	}

	t := persistence.GetTokenRepository()
	encKey := os.Getenv("ENCRYPTION_KEY")
	var userToken tokens.Token
	userToken.TokenType = fmt.Sprintf("Google-Calendar-Access-%s", tok.TokenType)
	userToken.UserID = user.ID
	userToken.HashedToken = crypto.EncryptString(tok.AccessToken, encKey)
	userToken.HashedRefreshToken = crypto.EncryptString(tok.RefreshToken, encKey)
	userToken.Expiry = tok.Expiry

	err = t.Add(&userToken)
	if err != nil {
		log.Printf("Unable to save token for user: %v", err)
	}

	return tok, nil
}

func GetCalendarInformation(ctx *gin.Context, tok *oauth2.Token) {
	//get user token, request calendar info

	b, err := os.ReadFile("calendar-credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, _ := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	client := getClientForUser(config, tok)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}

	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}

}
