package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eng_club/database"
)

// Bot handles Telegram bot commands and callbacks
type Bot struct {
	token   string
	db      *database.MySQLDatabaseManager
	baseURL string // Base URL for images (e.g., "http://localhost:8080")
}

// NewBot creates a new Telegram bot handler
func NewBot(token string, db *database.MySQLDatabaseManager, baseURL string) *Bot {
	return &Bot{
		token:   token,
		db:      db,
		baseURL: baseURL,
	}
}

// Update represents a Telegram update
type Update struct {
	UpdateID      int              `json:"update_id"`
	Message       *TelegramMessage `json:"message,omitempty"`
	CallbackQuery *CallbackQuery   `json:"callback_query,omitempty"`
}

// Message represents a Telegram message
type TelegramMessage struct {
	MessageID int    `json:"message_id"`
	From      User   `json:"from"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
	Date      int64  `json:"date"`
}

// User represents a Telegram user
type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// Chat represents a Telegram chat
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// CallbackQuery represents a callback query from an inline button
type CallbackQuery struct {
	ID      string           `json:"id"`
	From    User             `json:"from"`
	Message *TelegramMessage `json:"message,omitempty"`
	Data    string           `json:"data"`
}

// InlineKeyboardMarkup represents an inline keyboard
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton represents an inline keyboard button
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

// SendPhotoWithKeyboard sends a photo with caption and inline keyboard
func (b *Bot) SendPhotoWithKeyboard(chatID int64, photoURL string, caption string, keyboard *InlineKeyboardMarkup) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", b.token)

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"photo":        photoURL,
		"caption":      caption,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendMessageWithKeyboard sends a message with inline keyboard
func (b *Bot) SendMessageWithKeyboard(chatID int64, text string, keyboard *InlineKeyboardMarkup) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body[:n]))
	}

	return nil
}

// AnswerCallbackQuery answers a callback query
func (b *Bot) AnswerCallbackQuery(callbackQueryID, text string, showAlert bool) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", b.token)

	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
		"show_alert":        showAlert,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// EditMessageText edits a message text
func (b *Bot) EditMessageText(chatID int64, messageID int, text string, keyboard *InlineKeyboardMarkup) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", b.token)

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body[:n]))
	}

	return nil
}

// HandleStart handles the /start command
func (b *Bot) HandleStart(chatID int64, userID int64, username string) error {
	// Get next 6 planned events
	events, err := b.db.ListEvents("", "", "", "planned", 6, 0)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return b.SendMessageWithKeyboard(chatID, "No upcoming events scheduled yet. Check back later!", nil)
	}

	// Collect image filenames for collage
	var imageFilenames []string

	// Build the message text with all events
	messageText := "🎉 *Welcome to English Club!*\n\nHere are the upcoming events:\n\n"

	// Build inline keyboard with buttons for each event
	var buttons [][]InlineKeyboardButton

	for i, event := range events {
		// Check if user is already registered
		isRegistered, err := b.db.IsUserRegisteredForEvent(event.ID, userID)
		if err != nil {
			log.Printf("Error checking registration: %v", err)
		}

		// Get activity durations for this event
		totalDuration, err := b.db.GetActivityDurations(event.ID)
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
		if event.Place != nil {
			placeName = event.Place.Name
			if event.Place.MetroArea != "" {
				metroArea = event.Place.MetroArea
			}
			if event.Place.MapURL != "" {
				mapURL = event.Place.MapURL
			}
			// Collect image filename for collage
			if event.Place.ImageURL != "" {
				imageFilenames = append(imageFilenames, event.Place.ImageURL)
			} else {
				imageFilenames = append(imageFilenames, "")
			}
		} else {
			imageFilenames = append(imageFilenames, "")
		}

		// Add event details to message
		messageText += fmt.Sprintf("*%d. %s*%s\n", i+1, event.Topic, status)
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

		// Create button for this event
		buttonText := fmt.Sprintf("%d. %s", i+1, event.Topic)
		if isRegistered {
			buttonText = "✅ " + fmt.Sprintf("%d. %s", i+1, event.Topic)
		}

		// Add button to the keyboard (one button per row)
		buttons = append(buttons, []InlineKeyboardButton{
			{
				Text:         buttonText,
				CallbackData: fmt.Sprintf("event:%d:%s", event.ID, username),
			},
		})
	}

	messageText += "Select the event to register 👇"

	// Create the keyboard
	keyboard := &InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	// Create collage from place images
	// collageFilename, err := services.CreateCollage(imageFilenames)
	// if err != nil {
	// 	log.Printf("Failed to create collage: %v", err)
	// 	// Fallback to text-only message if collage creation fails
	// 	return b.SendMessageWithKeyboard(chatID, messageText, keyboard)
	// }

	// Send collage photo with message and buttons
	// collageURL := fmt.Sprintf("%s/uploads/collages/%s", b.baseURL, collageFilename)
	collageURL := "https://downloader.disk.yandex.ru/disk/b1292f7c03ab5fd8f559dc6b983c0fc380d37ff761e6ba8c5784bcb8d3da7a1d/69dd233a/8DOb7t76FtyaXH4kSFcDfPR3-ofg4a25I14M6phR-viRWqN1WcShRJxTbvotYrnlpIztyYy6ZSmfX2mjH-ehig%3D%3D?uid=0&filename=collage_787af0c8ee4ddd4139517be52033e980.jpg&disposition=attachment&hash=vYOi7J9RRc5S/f/Ze2MiJeZHOd0QqcsBaOQkqnrfF2DJcy%2Bzp2ye6CgiQzfXeyzoq/J6bpmRyOJonT3VoXnDag%3D%3D%3A&limit=0&content_type=image%2Fjpeg&owner_uid=689715285&fsize=56282&hid=c2d976ece0cf58659a597ec7ede7615b&media_type=image&tknv=v3"

	// Only send photo if it's an HTTPS URL (Telegram requirement)
	if strings.HasPrefix(collageURL, "https://") {
		err = b.SendPhotoWithKeyboard(chatID, collageURL, messageText, keyboard)
		if err != nil {
			log.Printf("Error sending collage photo: %v", err)
			// Fallback to text message if photo fails
			return b.SendMessageWithKeyboard(chatID, messageText, keyboard)
		}
		return nil
	} else {
		log.Printf("Skipping collage photo: Telegram requires HTTPS URLs (got: %s)", collageURL)
		// Fallback to text-only message for HTTP URLs
		return b.SendMessageWithKeyboard(chatID, messageText, keyboard)
	}
}

// HandleCallback handles callback queries from inline buttons
func (b *Bot) HandleCallback(query *CallbackQuery) error {
	parts := strings.Split(query.Data, ":")
	if len(parts) < 3 || parts[0] != "event" {
		return b.AnswerCallbackQuery(query.ID, "Invalid action", true)
	}

	eventID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return b.AnswerCallbackQuery(query.ID, "Invalid event ID", true)
	}

	username := parts[2]
	chatID := query.From.ID

	// Check if user is registered
	isRegistered, err := b.db.IsUserRegisteredForEvent(eventID, chatID)
	if err != nil {
		return b.AnswerCallbackQuery(query.ID, "Error checking registration", true)
	}

	if isRegistered {
		// Cancel registration
		memberID, err := b.db.GetMemberIDByChatID(chatID)
		if err != nil {
			return b.AnswerCallbackQuery(query.ID, "Failed to get member ID", true)
		}

		err = b.db.RemoveEventParticipant(eventID, memberID)
		if err != nil {
			return b.AnswerCallbackQuery(query.ID, "Failed to cancel registration", true)
		}

		// Update the message
		b.HandleStart(query.Message.Chat.ID, chatID, username)
		return b.AnswerCallbackQuery(query.ID, "Registration cancelled!", false)
	}

	// Register user
	// First, upsert the member
	chatIDStr := strconv.FormatInt(chatID, 10)
	memberID, err := b.db.UpsertMember(query.From.FirstName, username, chatIDStr, nil)
	if err != nil {
		return b.AnswerCallbackQuery(query.ID, "Failed to register", true)
	}

	// Add to event participants
	err = b.db.AddEventParticipant(eventID, memberID, nil, false)
	if err != nil {
		return b.AnswerCallbackQuery(query.ID, "Failed to register", true)
	}

	// Update the message
	b.HandleStart(query.Message.Chat.ID, chatID, username)
	return b.AnswerCallbackQuery(query.ID, "Successfully registered!", false)
}

// GetUpdates gets updates from Telegram
func (b *Bot) GetUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", b.token, offset)

	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Result, nil
}

// Start starts the bot polling loop
func (b *Bot) Start() {
	log.Println("Starting Telegram bot...")
	offset := 0

	for {
		updates, err := b.GetUpdates(offset)
		if err != nil {
			log.Printf("Error getting updates: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, update := range updates {
			offset = update.UpdateID + 1

			// Handle commands
			if update.Message != nil && update.Message.Text != "" {
				if strings.HasPrefix(update.Message.Text, "/start") {
					username := update.Message.From.Username
					if username == "" {
						username = update.Message.From.FirstName
					}
					err := b.HandleStart(update.Message.Chat.ID, update.Message.From.ID, username)
					if err != nil {
						log.Printf("Error handling /start: %v", err)
					}
				}
			}

			// Handle callback queries
			if update.CallbackQuery != nil {
				err := b.HandleCallback(update.CallbackQuery)
				if err != nil {
					log.Printf("Error handling callback: %v", err)
				}
			}
		}
	}
}
