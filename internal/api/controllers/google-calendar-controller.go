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

	"github.com/gin-gonic/gin"
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
	// request user permission with your credential - redirect to google
	// store token
	// use token in request
	RequestPermission(ctx)
}

func getClientForUser(config *oauth2.Config, userID int) *http.Client {
	tok := &oauth2.Token{}
	// get the saved token for the user
	return config.Client(context.Background(), tok)
}

func RequestPermission(ctx *gin.Context) {
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

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	ctx.Redirect(302, authURL)
}

func HandleGoogleAuthorizeCalendar(ctx *gin.Context) {
	appUrl := os.Getenv("APP_URL")
	handleGoogleAuthorize(ctx)
	ctx.Redirect(200, appUrl)
}

func handleGoogleAuthorize(ctx *gin.Context) (error, bool) {
	b, _ := os.ReadFile("calendar-credentials.json")
	code := ctx.Query("code")

	if code == "" {
		err := errors.New("no code")
		return err, false
	}

	config, _ := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	fmt.Printf("%v", tok)
	//save token for logged in user
	// put tok in a user token
	return nil, true
}

func GetCalendarInformation(ctx *gin.Context) {
	//get user token, request calendar info

	b, err := os.ReadFile("calendar-credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, _ := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	client := getClientForUser(config, 1)

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

// Retrieve a token, saves the token, then returns the generated client.
// func getClient(config *oauth2.Config) *http.Client {
// 	// The file token.json stores the user's access and refresh tokens, and is
// 	// created automatically when the authorization flow completes for the first
// 	// time.
// 	tokFile := "token.json"
// 	tok, err := tokenFromFile(tokFile)
// 	if err != nil {
// 		tok = getTokenFromWeb(config)
// 		saveToken(tokFile, tok)
// 	}
// 	return config.Client(context.Background(), tok)
// }

// Request a token from the web, then returns the retrieved token.
// func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
// 	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
// 	fmt.Printf("Go to the following link in your browser then type the "+
// 		"authorization code: \n%v\n", authURL)

// 	var authCode string
// 	if _, err := fmt.Scan(&authCode); err != nil {
// 		log.Fatalf("Unable to read authorization code: %v", err)
// 	}

// 	tok, err := config.Exchange(context.TODO(), authCode)
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve token from web: %v", err)
// 	}
// 	return tok
// }

// Retrieves a token from a local file.
// func tokenFromFile(file string) (*oauth2.Token, error) {
// 	f, err := os.Open(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer f.Close()
// 	tok := &oauth2.Token{}
// 	err = json.NewDecoder(f).Decode(tok)
// 	return tok, err
// }

// Saves a token to a file path.
// func saveToken(path string, token *oauth2.Token) {
// 	fmt.Printf("Saving credential file to: %s\n", path)
// 	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
// 	if err != nil {
// 		log.Fatalf("Unable to cache oauth token: %v", err)
// 	}
// 	defer f.Close()
// 	json.NewEncoder(f).Encode(token)
// }

// func ConnectToCalendar(ctx *gin.Context) {
// 	b, err := os.ReadFile("calendar-credentials.json")
// 	if err != nil {
// 		log.Fatalf("Unable to read client secret file: %v", err)
// 	}

// 	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
// 	if err != nil {
// 		log.Fatalf("Unable to parse client secret file to config: %v", err)
// 	}
// 	client := getClient(config)

// 	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve Calendar client: %v", err)
// 	}

// 	t := time.Now().Format(time.RFC3339)
// 	events, err := srv.Events.List("primary").ShowDeleted(false).
// 		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
// 	}
// 	fmt.Println("Upcoming events:")
// 	if len(events.Items) == 0 {
// 		fmt.Println("No upcoming events found.")
// 	} else {
// 		for _, item := range events.Items {
// 			date := item.Start.DateTime
// 			if date == "" {
// 				date = item.Start.Date
// 			}
// 			fmt.Printf("%v (%v)\n", item.Summary, date)
// 		}
// 	}
// }
