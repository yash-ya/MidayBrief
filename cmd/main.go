package main

import (
	"MidayBrief/internal/router"
	"MidayBrief/pkg/config"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.LoadEnv()
	router.SetupRoutes()

	port := ":8080"
	fmt.Println("Listening on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}