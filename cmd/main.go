package main

import (
	"MidayBrief/db"
	"MidayBrief/internal/router"
	"MidayBrief/pkg/config"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.LoadEnv()
	db.Init()

	err := db.DB.AutoMigrate(&db.TeamConfig{}, &db.UserMessage{})
	if err != nil {
        log.Fatal("❌ Auto migration failed:", err)
    }
    log.Println("✅ DB auto migration done")

	router.SetupRoutes()

	port := ":8080"
	fmt.Println("Listening on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}