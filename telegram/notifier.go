package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"eng_club/models"
)

// Notifier handles Telegram bot communications
type Notifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

// Message represents a message to send via Telegram
type Message struct {
	ChatID    *string `json:"chat_id"`
	Text      string  `json:"text"`
	ParseMode string  `json:"parse_mode,omitempty"`
}

// NewNotifier creates a new Telegram notifier
func NewNotifier(botToken, chatID string) *Notifier {
	return &Notifier{
		botToken: botToken,
		chatID:   chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendMessage sends a message to a Telegram chat
func (tn *Notifier) SendMessage(chatID *string, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tn.botToken)

	msg := Message{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := tn.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the error response body
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body[:n]))
	}

	return nil
}

func (tn *Notifier) NotifyTableAssignments(member models.Participant, tableID int) error {
	// Skip if member doesn't have a chat ID
	if member.TelegramChatID == nil {
		return nil
	}

	// Send individual message
	message := fmt.Sprintf(`🔄 *Your new table: %d*`, tableID)

	err := tn.SendMessage(member.TelegramChatID, message)
	if err != nil {
		// Log error but continue sending to other members
		return fmt.Errorf("Failed to send notification to %s: %v\n", member.Name, err)
	}

	return nil
}

// NotifyActivityEnd sends notifications when an activity is about to end
func (tn *Notifier) NotifyActivityEnd(activityName string, participants map[int]models.Participant, minutesRemaining int) error {
	for _, p := range participants {
		if p.TelegramChatID != nil {
			personalMessage := fmt.Sprintf(`⏰ *%s Ending Soon!*
🚨 Only %d minute remaining!
Please wrap up your conversation and prepare for the next activity.`, activityName, minutesRemaining)
			tn.SendMessage(p.TelegramChatID, personalMessage)
		}
	}

	return nil
}

// NotifySessionStart sends welcome messages when a session begins
func (tn *Notifier) NotifySessionStart(eventID int64, participants []models.Participant, duration int) error {
	// Send individual messages only to participants with chat IDs
	for _, p := range participants {
		if p.TelegramChatID != nil {
			personalMessage := fmt.Sprintf(`🎉 *Welcome to English Club, %s!*

You're registered for today's session!

📅 Event #%d
⏰ Total duration: ~%d minutes

Get ready for an amazing practice session! 🗣️✨`, p.Name, eventID, duration)

			if err := tn.SendMessage(p.TelegramChatID, personalMessage); err != nil {
				// Log error but continue with other participants
				fmt.Printf("Failed to send to %s (%v): %v\n", p.Name, p.TelegramChatID, err)
			}
		}
	}

	return nil
}

// NotifySessionEnd sends farewell messages when a session ends
func (tn *Notifier) NotifySessionEnd(sessionID int64, participants map[int]models.Participant) error {
	if tn.chatID == "" {
		return nil // Notifications disabled
	}

	for _, p := range participants {
		if p.TelegramChatID != nil {
			personalMessage := fmt.Sprintf(`🏁 *Session Complete!*
Rate today's session: /feedback\_%d`, sessionID)

			if err := tn.SendMessage(p.TelegramChatID, personalMessage); err != nil {
				// Log error but continue with other participants
				fmt.Printf("Failed to send to %s (%v): %v\n", p.Name, p.TelegramChatID, err)
			}
		}
	}

	return nil
}

// ActivitySchedule represents a scheduled activity with timing
type ActivitySchedule struct {
	ID           int64
	Name         string
	StartTime    time.Time
	EndTime      time.Time
	Participants []models.Participant
}

// NotifyUpcomingEvent sends notifications for an upcoming event to the general chat
func (tn *Notifier) NotifyUpcomingEvent(event models.Event, participants []models.Participant, totalDuration int) error {
	scheduledTime := event.ScheduledAt.Format("Mon, Jan 02 at 15:04")

	// Build general message text
	messageText := fmt.Sprintf(`🎉 *Reminder: English Club Event Today!*

*%s*

📅 %s
🕐 %d minutes`, event.Topic, scheduledTime, totalDuration)

	// Add place information if available
	if event.Place != nil {
		if event.Place.MapURL != "" {
			messageText += fmt.Sprintf("\n📍 [%s](%s)", event.Place.Name, event.Place.MapURL)
		} else {
			messageText += fmt.Sprintf("\n📍 %s", event.Place.Name)
		}

		if event.Place.MetroArea != "" {
			messageText += fmt.Sprintf(" 🚇 %s", event.Place.MetroArea)
		}
	}

	// Add participant count
	messageText += fmt.Sprintf("\n👥 %d participants registered", len(participants))
	messageText += "\n\nGet ready for an amazing practice session! 🗣️✨"

	// Send to notify_chat_id from config
	if tn.chatID != "" {
		if err := tn.SendMessage(&tn.chatID, messageText); err != nil {
			return fmt.Errorf("failed to send to notify chat: %v", err)
		}
	} else {
		return fmt.Errorf("notify_chat_id is not configured")
	}

	return nil
}
