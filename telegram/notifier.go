package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"eng_club/database"
	"eng_club/models"
)

// Notifier handles Telegram bot communications
type Notifier struct {
	*TelegramClient
	chatID  string
	baseURL string
	db      *database.MySQLDatabaseManager
}

// Message represents a message to send via Telegram
type Message struct {
	ChatID    *string `json:"chat_id"`
	Text      string  `json:"text"`
	ParseMode string  `json:"parse_mode,omitempty"`
}

// NewNotifier creates a new Telegram notifier
func NewNotifier(botToken, chatID, baseURL string, db *database.MySQLDatabaseManager) *Notifier {
	return &Notifier{
		TelegramClient: NewTelegramClient(botToken),
		chatID:         chatID,
		baseURL:        baseURL,
		db:             db,
	}
}

// SendMessage sends a message to a Telegram chat
func (tn *Notifier) SendMessage(chatID *string, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tn.TelegramClient.token)

	msg := Message{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := tn.TelegramClient.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
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

// NotifyUpcomingEvent sends notifications for an upcoming event to the general chat with place image and registration button
func (tn *Notifier) NotifyUpcomingEvent(event models.Event, totalDuration int) error {
	// Use the shared FormatEventInfo function with custom title for notifications
	customTitle := "🎉 *Reminder: English Club Event Today!*"
	eventInfo, err := FormatEventInfo(tn.db, &event, int64(0), 0, "notification", customTitle) // userID=0 for notifications, index=0, username="notification"
	if err != nil {
		return fmt.Errorf("error formatting event info: %v", err)
	}

	// Add the call-to-action message
	messageText := eventInfo.FormattedText + "Click the button below to register! 🗣️✨"

	// Create registration button
	keyboard := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{
					Text:         "📝 Register for Event",
					CallbackData: fmt.Sprintf("event:%d:notification", event.ID),
				},
			},
		},
	}

	// Send to notify_chat_id from config
	if tn.chatID == "" {
		return fmt.Errorf("notify_chat_id is not configured")
	}

	// Try to send with place image if available
	if eventInfo.ImageURL != "" {
		fullImageURL := fmt.Sprintf("%s/uploads/%s", tn.baseURL, eventInfo.ImageURL)
		err := tn.TelegramClient.SendPhotoWithKeyboard(tn.chatID, fullImageURL, messageText, keyboard)
		if err != nil {
			// Fallback to text message if photo fails
			return tn.TelegramClient.SendMessageWithKeyboard(tn.chatID, messageText, keyboard)
		}
		return nil
	} else {
		// Send text message with keyboard if no image
		return tn.TelegramClient.SendMessageWithKeyboard(tn.chatID, messageText, keyboard)
	}
}
