package main

import (
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

	// Create bot - use domain URL from config
	if cfg.Telegram.DomainURL == "" {
		log.Fatal("domain_url is required in telegram config for image serving")
	}
	bot := telegram.NewBot(cfg.Telegram.BotToken, db, cfg.Telegram.DomainURL)

	// Start bot based on configuration
	if cfg.Telegram.WebhookURL != "" && cfg.Telegram.WebhookPort > 0 {
		// Use webhook mode
		log.Printf("Starting Telegram bot in webhook mode on port %d", cfg.Telegram.WebhookPort)
		if err := bot.StartWebhookServer(cfg.Telegram.WebhookPort, cfg.Telegram.WebhookURL); err != nil {
			log.Fatalf("Failed to start webhook server: %v", err)
		}
	} else {
		// Fall back to polling mode
		log.Println("Telegram bot started in polling mode!")
		bot.Start() // This blocks
	}
}
