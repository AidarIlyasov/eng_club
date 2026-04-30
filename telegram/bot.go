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
	*TelegramClient
	db      *database.MySQLDatabaseManager
	baseURL string // Base URL for images (e.g., "http://localhost:8080")
}

// NewBot creates a new Telegram bot handler
func NewBot(token string, db *database.MySQLDatabaseManager, baseURL string) *Bot {
	return &Bot{
		TelegramClient: NewTelegramClient(token),
		db:             db,
		baseURL:        baseURL,
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

// AnswerCallbackQuery answers a callback query
func (b *Bot) AnswerCallbackQuery(callbackQueryID, text string, showAlert bool) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", b.TelegramClient.token)

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
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", b.TelegramClient.token)

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
		eventInfo, err := FormatEventInfo(b.db, &event, userID, i+1, username, "")
		if err != nil {
			log.Printf("Error formatting event info: %v", err)
			continue
		}

		// Add to message text
		messageText += eventInfo.FormattedText

		// Collect image filename for collage
		if eventInfo.ImageURL != "" {
			imageFilenames = append(imageFilenames, eventInfo.ImageURL)
		} else {
			imageFilenames = append(imageFilenames, "")
		}

		// Add button to the keyboard (one button per row)
		buttons = append(buttons, []InlineKeyboardButton{
			{
				Text:         eventInfo.ButtonText,
				CallbackData: eventInfo.CallbackData,
			},
		})
	}

	messageText += "Select the event to register 👇"

	// Create the keyboard
	keyboard := &InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	// Create collage from place images
	collageFilename, err := CreateCollage(imageFilenames)
	if err != nil {
		log.Printf("Failed to create collage: %v", err)
		// Fallback to text-only message if collage creation fails
		return b.SendMessageWithKeyboard(chatID, messageText, keyboard)
	}

	// Send collage photo with message and buttons
	collageURL := fmt.Sprintf("%s/api/uploads/collages/%s", b.baseURL, collageFilename)

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

	var username string
	if parts[2] == "notification" {
		// Handle callback from notification - get username from query
		username = query.From.Username
		if username == "" {
			username = query.From.FirstName
		}
	} else {
		username = parts[2]
	}
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

		// Send formatted event info showing updated status with photo
		event, _ := b.db.GetEvent(eventID)
		if event != nil {
			eventInfo, _ := FormatEventInfo(b.db, event, chatID, 0, username, "✅ *Registration Cancelled*")
			if eventInfo.ImageURL != "" {
				fullImageURL := fmt.Sprintf("%s/api/uploads/%s", b.baseURL, eventInfo.ImageURL)
				err := b.TelegramClient.SendPhotoWithKeyboard(chatID, fullImageURL, eventInfo.FormattedText, nil)
				if err != nil {
					// Fallback to text message if photo fails
					b.TelegramClient.SendMessageWithKeyboard(chatID, eventInfo.FormattedText, nil)
				}
			} else {
				b.TelegramClient.SendMessageWithKeyboard(chatID, eventInfo.FormattedText, nil)
			}
		}

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

	// Send formatted event info showing updated status with photo
	event, _ := b.db.GetEvent(eventID)
	if event != nil {
		eventInfo, _ := FormatEventInfo(b.db, event, chatID, 0, username, "🎉 *Registration Successful*")
		if eventInfo.ImageURL != "" {
			fullImageURL := fmt.Sprintf("%s/api/uploads/%s", b.baseURL, eventInfo.ImageURL)
			err := b.TelegramClient.SendPhotoWithKeyboard(chatID, fullImageURL, eventInfo.FormattedText, nil)
			if err != nil {
				// Fallback to text message if photo fails
				b.TelegramClient.SendMessageWithKeyboard(chatID, eventInfo.FormattedText, nil)
			}
		} else {
			b.TelegramClient.SendMessageWithKeyboard(chatID, eventInfo.FormattedText, nil)
		}
	}

	return b.AnswerCallbackQuery(query.ID, "Successfully registered!", false)
}

// GetUpdates gets updates from Telegram
func (b *Bot) GetUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", b.TelegramClient.token, offset)

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

// ProcessUpdate processes a single Telegram update
func (b *Bot) ProcessUpdate(update Update) {
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

// HandleWebhook handles incoming webhook requests from Telegram
func (b *Bot) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Error decoding webhook update: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process the update
	b.ProcessUpdate(update)

	// Send OK response to Telegram
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// SetWebhook sets the webhook URL with Telegram
func (b *Bot) SetWebhook(webhookURL string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", b.TelegramClient.token)

	payload := map[string]interface{}{
		"url": webhookURL,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := b.TelegramClient.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Webhook set successfully to: %s", webhookURL)
	return nil
}

// StartWebhookServer starts the webhook HTTP server
func (b *Bot) StartWebhookServer(port int, webhookURL string) error {
	// Set webhook with Telegram
	if err := b.SetWebhook(webhookURL); err != nil {
		return fmt.Errorf("failed to set webhook: %v", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", b.HandleWebhook)

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	log.Printf("Starting Telegram webhook server on port %d", port)
	return server.ListenAndServe()
}

// Start starts the bot polling loop (legacy mode)
func (b *Bot) Start() {
	log.Println("Starting Telegram bot in polling mode...")
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
			b.ProcessUpdate(update)
		}
	}
}
