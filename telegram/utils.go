package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TelegramClient provides common Telegram API functionality
type TelegramClient struct {
	token  string
	client *http.Client
}

// NewTelegramClient creates a new Telegram client
func NewTelegramClient(token string) *TelegramClient {
	return &TelegramClient{
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendPhotoWithKeyboard sends a photo with caption and inline keyboard
func (tc *TelegramClient) SendPhotoWithKeyboard(chatID interface{}, photoURL string, caption string, keyboard *InlineKeyboardMarkup) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", tc.token)

	payload := PhotoMessage{
		ChatID:      chatID,
		Photo:       photoURL,
		Caption:     caption,
		ParseMode:   "Markdown",
		ReplyMarkup: keyboard,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := tc.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body[:n]))
	}

	return nil
}

// SendMessageWithKeyboard sends a message with inline keyboard
func (tc *TelegramClient) SendMessageWithKeyboard(chatID interface{}, text string, keyboard *InlineKeyboardMarkup) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tc.token)

	payload := MessageWithKeyboard{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   "Markdown",
		ReplyMarkup: keyboard,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := tc.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body[:n]))
	}

	return nil
}
