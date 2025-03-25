package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
)

// StartHandler handles the /start endpoint
func StartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clock, err := loadFromFile()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error loading file: %v", err), http.StatusInternalServerError)
			return
		}
		clock.start()

		response := ClockResponse{
			TotalSeconds: clock.TotalWorkedSec,
		}

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		json.NewEncoder(w).Encode(response)
	}
}

// PauseHandler handles the /pause endpoint
func PauseHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clock, err := loadFromFile()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error loading file: %v", err), http.StatusInternalServerError)
			return
		}
		clock.pause()

		response := ClockResponse{
			TotalSeconds: clock.TotalWorkedSec,
		}

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		json.NewEncoder(w).Encode(response)
	}
}

// StatusHandler handles the /status endpoint
func StatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clock, err := loadFromFile()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error loading file: %v", err), http.StatusInternalServerError)
			return
		}
		var elapsedSeconds int64 = 0
		if clock.IsRunning {
			elapsedSeconds = int64(time.Since(clock.StartTime).Seconds())
		}

		response := StatusResponse{
			IsRunning:      clock.IsRunning,
			TotalSeconds:   clock.TotalWorkedSec,
			ElapsedSeconds: elapsedSeconds,
			Filename:       Filename(),
		}

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		json.NewEncoder(w).Encode(response)
	}
}
