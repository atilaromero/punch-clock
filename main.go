package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
)

type PunchClock struct {
	IsRunning      bool
	StartTime      time.Time
	TotalWorkedSec int64
	LastSaveTime   time.Time
	Filename       string
}

func (p *PunchClock) start() {
	if !p.IsRunning {
		p.StartTime = time.Now()
		p.IsRunning = true
		p.saveRecord("start")
	}
}

func (p *PunchClock) pause() {
	if p.IsRunning {
		p.TotalWorkedSec += int64(time.Since(p.StartTime).Seconds())
		p.IsRunning = false
		p.saveRecord("pause")
	}
}

func (p *PunchClock) getCurrentWorkedTime() (hours, minutes, seconds int) {
	var totalSec int64 = int64(p.TotalWorkedSec)
	if p.IsRunning {
		totalSec += int64(time.Since(p.StartTime).Seconds())
	}
	hours = int(totalSec / 3600)
	minutes = int((totalSec % 3600) / 60)
	seconds = int(totalSec % 60)
	return
}

func (p *PunchClock) formatWorkedTime() string {
	hours, minutes, seconds := p.getCurrentWorkedTime()
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func (p *PunchClock) saveRecord(action string) {
	// Create the file if it doesn't exist
	file, err := os.OpenFile(p.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the record
	record := []string{
		action,
		p.formatWorkedTime(),
		time.Now().Format(time.RFC3339),
	}
	if err := writer.Write(record); err != nil {
		log.Printf("Error writing record: %v", err)
	}
	p.LastSaveTime = time.Now()
}

func (p *PunchClock) loadFromFile() {
	file, err := os.Open(p.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("No existing file for today, starting fresh")
			return
		}
		log.Printf("Error opening file: %v", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = 3

	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV: %v", err)
		return
	}

	if len(records) == 0 {
		return
	}

	// Process the last record to determine the current state
	lastRecord := records[len(records)-1]
	action := lastRecord[0]
	timeStr := lastRecord[1]
	
	// Parse the last worked time
	parts := strings.Split(timeStr, ":")
	if len(parts) == 3 {
		hours, _ := strconv.Atoi(parts[0])
		minutes, _ := strconv.Atoi(parts[1])
		seconds, _ := strconv.Atoi(parts[2])
		p.TotalWorkedSec = int64(hours*3600 + minutes*60 + seconds)
	}

	// Set the current state based on the last action
	p.IsRunning = action == "start"
	if p.IsRunning {
		// If the clock was running, we need to account for the time since the last record
		timestamp, err := time.Parse(time.RFC3339, lastRecord[2])
		if err == nil {
			p.StartTime = timestamp
		} else {
			p.StartTime = time.Now()
		}
	}
}

// StatusResponse represents the JSON response for status endpoint
type StatusResponse struct {
	IsRunning      bool   `json:"isRunning"`
	TotalSeconds   int64  `json:"totalSeconds"`
	ElapsedSeconds int64  `json:"elapsedSeconds"`
	Filename       string `json:"filename"`
}

// ClockResponse represents the JSON response for start/pause endpoints
type ClockResponse struct {
	TotalSeconds int64 `json:"totalSeconds"`
}

func main() {
	// Initialize the punch clock
	today := time.Now().Format("2006-01-02")
	filename := today + ".csv"
	
	clock := &PunchClock{
		IsRunning:      false,
		TotalWorkedSec: 0,
		Filename:       filename,
	}
	
	// Load existing data if available
	clock.loadFromFile()
	
	// Set up HTTP handlers
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	
	// Start endpoint
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		clock.start()
		
		response := ClockResponse{
			TotalSeconds: clock.TotalWorkedSec,
		}
		
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		json.NewEncoder(w).Encode(response)
	})
	
	// Pause endpoint
	http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		clock.pause()
		
		response := ClockResponse{
			TotalSeconds: clock.TotalWorkedSec,
		}
		
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		json.NewEncoder(w).Encode(response)
	})
	
	// Status endpoint for initial page load
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		var elapsedSeconds int64 = 0
		if clock.IsRunning {
			elapsedSeconds = int64(time.Since(clock.StartTime).Seconds())
		}
		
		response := StatusResponse{
			IsRunning:      clock.IsRunning,
			TotalSeconds:   clock.TotalWorkedSec,
			ElapsedSeconds: elapsedSeconds,
			Filename:       filename,
		}
		
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		json.NewEncoder(w).Encode(response)
	})
	
	// Start the server
	port := 8080
	fmt.Printf("Starting punch clock server on http://localhost:%d\n", port)
	fmt.Printf("Using data file: %s\n", filename)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
