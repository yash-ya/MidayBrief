package main

import (
	"log"
	"net/http"
	"os"

	"MidayBrief/db"
	"MidayBrief/scheduler"
	"MidayBrief/utils"
)

func main() {
	db.Init()
	utils.InitCrypto()
	go scheduler.StartScheduler()
	router := SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server failed:", err)
	}
}
