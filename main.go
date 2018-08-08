package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type eventdata struct {
	// Date: The date, in the format "yyyy-mm-dd", if this is an all-day
	// event.
	time    time.Time
	message string
}

func (instance eventdata) isAfter(t time.Time) bool {
	return instance.time.After(t)
}

// Define custom array type that implements Sort
type eventsList []eventdata

func (instance eventsList) Len() int {
	return len(instance)
}
func (instance eventsList) Less(i, j int) bool {
	return instance[i].time.Before(instance[j].time)
}
func (instance eventsList) Swap(i, j int) {
	instance[i], instance[j] = instance[j], instance[i]
}
func (instance eventsList) contains(event eventdata) bool {
	for _, e := range instance {
		if event.time == e.time && event.message == e.message {
			return true
		}
	}
	return false
}

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	// Setup
	now := time.Now()
	// times formatted as "2006-01-02T15:04:05Z07:00"
	t := now.Format(time.RFC3339)
	tomorrow := now.AddDate(0, 0, 1).Format(time.RFC3339) // add a week

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	// loop through all calendars and get their names
	calendars, err := srv.CalendarList.List().Do()
	if err != nil {
		log.Fatalf("unable to grab all calendars")
	}

	// grab all of the events into array
	allEvents := make(eventsList, 0)

	for _, cal := range calendars.Items {
		events, err := srv.Events.List(cal.Id).ShowDeleted(false).SingleEvents(true).TimeMin(t).TimeMax(tomorrow).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve for calendar %v: %v", cal.Summary, err)
		}

		if len(events.Items) == 0 {
			fmt.Printf("No upcoming events found for calendar %v\n", cal.Summary)
		} else {
			// fmt.Println("Upcoming events for calendar %v:", cal.Id)
			for _, item := range events.Items {
				if item.Start != nil {
					date := item.Start.DateTime
					if date == "" {
						date = item.Start.Date
					}

					t, _ := time.Parse(time.RFC3339, date)
					reformattedTime := t.Format(time.RFC1123)
					msg := fmt.Sprintf("%v -- %v\n", reformattedTime, item.Summary)
					newEvent := eventdata{
						time:    t,
						message: msg}

					// skip dupes
					if !allEvents.contains(newEvent) {
						allEvents = append(allEvents, newEvent)
					}
				}
			}
		}
	}

	// sort the events
	sort.Sort(allEvents)

	beforeEvents := make(eventsList, len(allEvents))
	afterEvents := make(eventsList, len(allEvents))

	// loop through all events!
	for _, event := range allEvents {
		if event.isAfter(now) {
			afterEvents = append(afterEvents, event)
		} else {
			beforeEvents = append(beforeEvents, event)
		}
	}

	fmt.Printf("\n\n ===== ONGOING ===== \n\n")

	for _, event := range beforeEvents {
		fmt.Printf(event.message)
	}

	fmt.Printf("\n\n ===== TODAY ===== \n\n")

	for _, event := range afterEvents {
		fmt.Printf(event.message)
	}

}

// ===== Oauth Stuff =====

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
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

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}
