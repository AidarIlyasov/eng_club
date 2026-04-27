package handlers

import (
	"eng_club/database"
	"eng_club/telegram"
)

// Deps holds shared dependencies for HTTP handlers.
type Deps struct {
	DB          *database.MySQLDatabaseManager
	TelegramBot *telegram.Notifier
}
