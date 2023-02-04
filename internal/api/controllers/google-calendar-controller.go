package controllers

import (
	"context"
	"encoding/json"
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

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func ConnectToCalendar(ctx *gin.Context) {
	b, err := os.ReadFile("calendar-credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// cid := os.Getenv("GOOGLE_CLIENT_ID")
	// secret := os.Getenv("GOOGLE_OAUTH_SECRET")

	// config := &oauth2.Config{ClientID: cid, ClientSecret: secret }

	// token, err := config.Exchange(ctx, code)

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

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
	// //  credential := ctx.PostForm("credential")

	// // if credential == "" {
	// // 	ctx.JSON(http.StatusBadRequest, "")
	// // }

	// code := ctx.Request.URL.Query().Get("code")

	// cid := os.Getenv("GOOGLE_CLIENT_ID")
	// secret := os.Getenv("GOOGLE_OAUTH_SECRET")

	// config := &oauth2.Config{ClientID: cid, ClientSecret: secret }

	// token, err := config.Exchange(ctx, code)

	// //  config, err := google.ConfigFromJSON([]byte(credential), calendar.CalendarReadonlyScope)
	// //  if err != nil {
	// // 	 log.Fatalf("Unable to parse client secret file to config: %v", err)
	// //  }

	//  // Redirect user to Google's OAuth2 consent page to grant the application access to their calendar

	//  if err != nil {
	// 	 log.Fatalf("Unable to retrieve token from web: %v", err)
	//  }

	//  calendar, err := calendar.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	//  if err != nil {
	// 	log.Fatalf("Unable to retrieve calendar client %v", err)
	// }

	// //  client := config.Client(ctx, token)
	// //  srv, err := calendar.New(client)
	// //  if err != nil {
	// // 	 log.Fatalf("Unable to retrieve calendar client %v", err)
	// //  }

	//  calendarRes, err := calendar.Calendars.Get("primary").Do()
	//  if err != nil {
	// 	 log.Fatalf("Unable to retrieve calendar: %v", err)
	//  }

	//  // Store the calendar summary in a JSON object
	//  calendarJSON, err := json.Marshal(calendarRes)
	//  if err != nil {
	// 	 log.Fatalf("Unable to parse calendar to JSON: %v", err)
	//  }

	//  fmt.Fprintf(ctx.Writer, string(calendarJSON))

}
