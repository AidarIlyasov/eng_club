package telegram

import (
	"fmt"
	"log"

	"eng_club/models"
	"eng_club/database"
)

// EventInfo holds formatted event information
type EventInfo struct {
	FormattedText    string
	ButtonText       string
	CallbackData     string
	ImageURL         string
	TotalDuration    int
	IsRegistered     bool
}

// FormatEventInfo formats event information for display
func FormatEventInfo(db *database.MySQLDatabaseManager, event *models.Event, userID int64, index int, username string, customTitle string) (*EventInfo, error) {
	// Check if user is already registered
	isRegistered, err := db.IsUserRegisteredForEvent(event.ID, userID)
	if err != nil {
		log.Printf("Error checking registration: %v", err)
	}

	// Get activity durations for this event
	totalDuration, err := db.GetActivityDurations(event.ID)
	if err != nil {
		log.Printf("Error getting durations for event %d: %v", event.ID, err)
	}

	// Format event info
	scheduledTime := event.ScheduledAt.Format("Mon, Jan 02 at 15:04")
	status := ""
	if isRegistered {
		status = " ✅ - Registered"
	}

	placeName := "Unknown"
	metroArea := ""
	mapURL := ""
	imageURL := ""
	if event.Place != nil {
		placeName = event.Place.Name
		if event.Place.MetroArea != "" {
			metroArea = event.Place.MetroArea
		}
		if event.Place.MapURL != "" {
			mapURL = event.Place.MapURL
		}
		if event.Place.ImageURL != "" {
			imageURL = event.Place.ImageURL
		}
	}

	// Build formatted text
	var messageText string
	if customTitle != "" {
		messageText = fmt.Sprintf("%s\n\n*%s*%s\n", customTitle, event.Topic, status)
	} else if index > 0 {
		messageText = fmt.Sprintf("*%d. %s*%s\n", index, event.Topic, status)
	} else {
		messageText = fmt.Sprintf("*%s*%s\n", event.Topic, status)
	}

	// Make place name a clickable link if map URL is available
	if mapURL != "" {
		messageText += fmt.Sprintf("📍 [%s](%s)", placeName, mapURL)
	} else {
		messageText += fmt.Sprintf("📍 %s", placeName)
	}
	if metroArea != "" {
		messageText += fmt.Sprintf(" 🚇 %s", metroArea)
	}
	messageText += fmt.Sprintf("\n📅 %s", scheduledTime)
	messageText += fmt.Sprintf(" 🕐 %d min\n\n", totalDuration)

	// Create button text
	var buttonText string
	if index > 0 {
		buttonText = fmt.Sprintf("%d. %s", index, event.Topic)
		if isRegistered {
			buttonText = "✅ " + fmt.Sprintf("%d. %s", index, event.Topic)
		}
	} else {
		buttonText = "📝 Register for Event"
		if isRegistered {
			buttonText = "❌ Cancel Registration"
		}
	}

	// Create callback data
	callbackData := fmt.Sprintf("event:%d:%s", event.ID, username)

	return &EventInfo{
		FormattedText:    messageText,
		ButtonText:       buttonText,
		CallbackData:     callbackData,
		ImageURL:         imageURL,
		TotalDuration:    totalDuration,
		IsRegistered:     isRegistered,
	}, nil
}