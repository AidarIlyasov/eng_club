package main

import (
	"log"
	"os"

	"eng_club/config"
	"eng_club/database"
)

func main() {
	cfgPath := ""
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	cfg, err := config.LoadAppConfig(cfgPath)
	if err != nil {
		log.Fatal("config: ", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal("invalid config: ", err)
	}
	if err := database.RunMigrations(cfg.MySQL()); err != nil {
		log.Fatal("migrate: ", err)
	}
	log.Println("migrations applied successfully")
}
