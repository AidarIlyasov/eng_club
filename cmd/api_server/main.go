package main

import (
	"log"
	"math/rand"
	"time"

	"eng_club/api"
	"eng_club/config"
	"eng_club/database"
	"eng_club/services"
	"eng_club/telegram"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Set global timezone to Moscow
	moscowTZ, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatal("Failed to load Moscow timezone: ", err)
	}
	time.Local = moscowTZ

	cfg, err := config.LoadAppConfig("")
	if err != nil {
		log.Fatal("config: ", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal("invalid config: ", err)
	}

	db, err := database.NewMySQLDatabaseManager(cfg.MySQL())
	if err != nil {
		log.Fatal("database: ", err)
	}

	if cfg.Telegram.DomainURL == "" {
		log.Fatal("domain_url is required in telegram config for image serving")
	}
	telegramBot := telegram.NewNotifier(cfg.Telegram.BotToken, cfg.Telegram.NotifyChatID, cfg.Telegram.DomainURL, db)

	// Create notification service and start scheduler
	notificationService := services.NewNotificationService(db, telegramBot)

	// Schedule notification checking every minute
	log.Println("Starting notification scheduler...")
	services.Schedule(60, func() {
		if err := notificationService.CheckAndSendEventNotifications(); err != nil {
			log.Printf("Error checking notifications: %v", err)
		}
	}, false) // Start immediately

	srv := api.NewServer(cfg, db, telegramBot)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
