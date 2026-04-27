package main

import (
	"fmt"
	"log"
	"time"

	"eng_club/config"
	"eng_club/database"
	"eng_club/telegram"
)

func main() {
	// Set global timezone to Moscow
	moscowTZ, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatal("Failed to load Moscow timezone: ", err)
	}
	time.Local = moscowTZ

	// Load configuration
	cfg, err := config.LoadAppConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	if !cfg.Telegram.Enabled {
		log.Fatal("Telegram bot is disabled in config")
	}

	// Initialize database
	db, err := database.NewMySQLDatabaseManager(cfg.MySQL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create and start bot
	baseURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	bot := telegram.NewBot(cfg.Telegram.BotToken, db, baseURL)
	log.Println("Telegram bot started successfully!")
	bot.Start() // This blocks
}
