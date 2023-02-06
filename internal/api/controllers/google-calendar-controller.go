package controllers

import (
	"context"

	// "encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/antonioalfa22/go-rest-template/internal/pkg/models/calendars"
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

const credentialsPath = "calendar-credentials.json"

type GoogleCalendar struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"required"`
	AccessRole      string `json:"accessRole,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Deleted         bool   `json:"deleted,omitempty"`
	Description     string `json:"description,omitempty"`
	Etag            string `json:"etag,omitempty"`
	ForegroundColor string `json:"foregroundColor,omitempty"`
	Hidden          bool   `json:"hidden,omitempty"`
	CalendarId      string `json:"id,omitempty"`
	Kind            string `json:"kind,omitempty"`
	Location        string `json:"location,omitempty"`
	Primary         bool   `json:"primary,omitempty"`
	Summary         string `json:"summary,omitempty"`
	TimeZone        string `json:"timeZone,omitempty"`
	CalendarType    string `json:"type,omitempty"`
}

type ConnectGoogleCalendarResponse struct {
	Url string `json:"url" binding:"required"`
}

type userCalendarResponseInfoDTO struct {
	userID  uuid.UUID
	tokenID uuid.UUID
	token   oauth2.Token
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
	userValue := *loggedInUser
	user, err := u.GetByEmail(userValue.User.Email)

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
	b, err := os.ReadFile(credentialsPath)
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
	encodedUUID := url.QueryEscape(userIDString)
	authURL := config.AuthCodeURL(encodedUUID, oauth2.AccessTypeOffline)

	var response ConnectGoogleCalendarResponse
	response.Url = authURL

	log.Printf("response url is returned %s", authURL)

	ctx.JSON(http.StatusOK, gin.H{"data": response})
}

func HandleGoogleAuthorizeCalendar(ctx *gin.Context) {
	appUrl := os.Getenv("APP_URL")
	token, userID, tokenID, err := handleGoogleAuthorize(ctx)

	if err != nil {
		log.Printf("unable to get user token %v", err)
		ctx.Redirect(302, appUrl)
	}

	var calendarInfo userCalendarResponseInfoDTO
	calendarInfo.token = *token
	calendarInfo.userID = *userID
	calendarInfo.tokenID = *tokenID

	GetCalendarInformation(ctx, calendarInfo)
	ctx.Redirect(302, appUrl+"/user/dashboard/calendar/connected")
}

func handleGoogleAuthorize(ctx *gin.Context) (*oauth2.Token, *uuid.UUID, *uuid.UUID, error) {
	b, _ := os.ReadFile(credentialsPath)
	code := ctx.Query("code")
	state := ctx.Query("state")
	decodedState, err := url.QueryUnescape(state)
	if err != nil {
		fmt.Printf("decoding state error %v", err)
	}

	if code == "" {
		err := errors.New("no code")
		return nil, nil, nil, err
	}

	if decodedState == "" {
		err := errors.New("no state parameter. invalid request")
		return nil, nil, nil, err
	}

	config, _ := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Printf("Unable to retrieve token from web: %v", err)
		return nil, nil, nil, err
	}

	u := persistence.GetUserRepository()
	userIDStr := crypto.DecryptString(decodedState, os.Getenv("ENCRYPTION_KEY"))
	userId, err := uuid.Parse(userIDStr)

	if err != nil {
		log.Printf("uuid parsing error: %v", err)
		return nil, nil, nil, err
	}

	user, err := u.Get(userId)
	if err != nil {
		return nil, nil, nil, err
	}
	// checkLogged in user is the same as the user for this request
	_, loggedIn, err := CheckUserThatIsLoggedIn((ctx))
	if err != nil {
		log.Printf("could not find a logged in user")
		return nil, nil, nil, err
	}

	//try to fetch logged in user
	luser, err := u.GetByEmail(loggedIn.User.Email)
	if err != nil {
		log.Printf("could not find a logged in user with this email")
		return nil, nil, nil, err
	}

	if luser.ID != userId {
		log.Printf("user mismatch!! malicious attempt may have been attempted")
		return nil, nil, nil, err
	}

	t := persistence.GetTokenRepository()
	encKey := os.Getenv("ENCRYPTION_KEY")
	var userToken tokens.Token
	userToken.TokenType = fmt.Sprintf("Google-Calendar-Access-%s", tok.TokenType)
	userToken.UserID = user.ID
	userToken.HashedToken = crypto.EncryptString(tok.AccessToken, encKey)
	userToken.HashedRefreshToken = crypto.EncryptString(tok.RefreshToken, encKey)
	userToken.Expiry = tok.Expiry
	userToken.ID = uuid.New()

	err = t.Add(&userToken)
	if err != nil {
		log.Printf("Unable to save token for user: %v", err)
	}

	return tok, &userId, &userToken.ID, nil
}

func GetCalendarInformation(ctx *gin.Context, calendarInfo userCalendarResponseInfoDTO) {
	// with the user token, request the calendar and save it.
	tok := calendarInfo.token

	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		log.Printf("Unable to read client secret file: %v", err)
		return
	}

	config, _ := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	client := getClientForUser(config, &tok)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Calendar client: %v", err)
		return
	}

	calendar, err := srv.CalendarList.Get("primary").Do()

	if err != nil {
		log.Printf("Unable to retrieve Primary Calendar for user, error occurred %v", err)
		return
	}

	if calendar == nil {
		log.Println("Unable to retrieve Primary Calendar for user")
		return
	}

	var calendarToSave calendars.Calendar
	calendarSummary := calendar.Summary
	if len(calendar.SummaryOverride) > 0 {
		calendarSummary = calendar.SummaryOverride
	}
	calendarToSave.AccessRole = calendar.AccessRole
	calendarToSave.CalendarId = calendar.Id
	calendarToSave.BackgroundColor = calendar.BackgroundColor
	calendarToSave.TimeZone = calendar.TimeZone
	calendarToSave.Deleted = calendar.Deleted
	calendarToSave.Description = calendar.Description
	calendarToSave.Kind = calendar.Kind
	calendarToSave.ForegroundColor = calendar.ForegroundColor
	calendarToSave.Location = calendar.Location
	calendarToSave.Summary = calendarSummary
	calendarToSave.Primary = calendar.Primary
	calendarToSave.CalendarType = "Google"
	calendarToSave.ID = uuid.New()
	calendarToSave.TokenID = calendarInfo.tokenID
	calendarToSave.UserID = calendarInfo.userID

	s := persistence.GetCalendarRepository()
	err = s.Add(&calendarToSave)
	if err != nil {
		log.Println("Unable to retrieve save calendar for user")
		return
	}

	// get and save primary calendar

	// t := time.Now().Format(time.RFC3339)
	// events, err := srv.Events.List("primary").ShowDeleted(false).
	// 	SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	// if err != nil {
	// 	log.Printf("Unable to retrieve next ten of the user's events: %v", err)
	// 	return
	// }

	// fmt.Println("Upcoming events:")
	// if len(events.Items) == 0 {
	// 	fmt.Println("No upcoming events found.")
	// } else {
	// 	for _, item := range events.Items {
	// 		date := item.Start.DateTime
	// 		if date == "" {
	// 			date = item.Start.Date
	// 		}
	// 		fmt.Printf("%v (%v)\n", item.Summary, date)
	// 	}
	// }

}
