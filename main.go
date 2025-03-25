package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Set up HTTP handlers
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// Register API endpoints
	http.HandleFunc("/start", StartHandler())
	http.HandleFunc("/pause", PauseHandler())
	http.HandleFunc("/status", StatusHandler())

	// Start the server
	port := 8080
	fmt.Printf("Starting punch clock server on http://localhost:%d\n", port)
	fmt.Printf("Using data file: %s\n", Filename())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
