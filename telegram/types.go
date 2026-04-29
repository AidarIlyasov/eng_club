package telegram

// Shared Telegram API types used by both Bot and Notifier

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

// PhotoMessage represents a photo message to send via Telegram
type PhotoMessage struct {
	ChatID      interface{}               `json:"chat_id"` // Can be string or int64
	Photo       string                    `json:"photo"`
	Caption     string                    `json:"caption"`
	ParseMode   string                    `json:"parse_mode,omitempty"`
	ReplyMarkup *InlineKeyboardMarkup     `json:"reply_markup,omitempty"`
}

// MessageWithKeyboard represents a text message with inline keyboard
type MessageWithKeyboard struct {
	ChatID      interface{}               `json:"chat_id"` // Can be string or int64
	Text        string                    `json:"text"`
	ParseMode   string                    `json:"parse_mode,omitempty"`
	ReplyMarkup *InlineKeyboardMarkup     `json:"reply_markup,omitempty"`
}