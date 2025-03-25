package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PunchClock represents the time tracking system
type PunchClock struct {
	IsRunning      bool
	StartTime      time.Time
	TotalWorkedSec int64
	LastSaveTime   time.Time
}

// Filename returns the CSV filename for the current day
func Filename() string {
	today := time.Now().Format("2006-01-02")
	return today + ".csv"
}

// Start begins a work session
func (p *PunchClock) start() {
	if !p.IsRunning {
		p.StartTime = time.Now()
		p.IsRunning = true
		p.saveRecord("start")
	}
}

// Pause stops the current work session
func (p *PunchClock) pause() {
	if p.IsRunning {
		p.TotalWorkedSec += int64(time.Since(p.StartTime).Seconds())
		p.IsRunning = false
		p.saveRecord("pause")
	}
}

// GetCurrentWorkedTime calculates the total time worked in hours, minutes, and seconds
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

// FormatWorkedTime returns a formatted string of the total time worked
func (p *PunchClock) formatWorkedTime() string {
	hours, minutes, seconds := p.getCurrentWorkedTime()
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// SaveRecord writes an action record to the CSV file
func (p *PunchClock) saveRecord(action string) {
	filename := Filename()
	// Create the file if it doesn't exist
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// LoadFromFile reads the current day's CSV file and updates the PunchClock state
func loadFromFile() (*PunchClock, error) {
	p := &PunchClock{
		IsRunning:      false,
		TotalWorkedSec: 0,
		StartTime:      time.Now(),
		LastSaveTime:   time.Now(),
	}
	filename := Filename()
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			p.TotalWorkedSec = 0
			p.IsRunning = false
			log.Printf("No existing file for today, starting fresh")
			return p, nil
		}
		log.Printf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = 3

	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV: %v", err)
		return nil, err
	}

	if len(records) == 0 {
		return p, nil
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

	return p, nil
}

// HistoryEntry represents a single day's work history
type HistoryEntry struct {
	Date  string  `json:"date"`
	Hours float64 `json:"hours"`
}

// GetHistoryData reads all CSV files in the current directory and returns the history data
func GetHistoryData() ([]HistoryEntry, error) {
	files, err := os.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var history []HistoryEntry
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, ".csv") {
			continue
		}

		// Try to parse the date from the filename (format: YYYY-MM-DD.csv)
		date := strings.TrimSuffix(name, ".csv")
		if len(date) != 10 || date[4] != '-' || date[7] != '-' {
			continue // Skip files that don't match the expected date format
		}

		// Read the file to get the total hours worked
		entry, err := getHistoryEntryFromFile(name, date)
		if err != nil {
			log.Printf("Error processing file %s: %v", name, err)
			continue
		}

		history = append(history, entry)
	}

	// Sort history by date (newest first)
	sort.Slice(history, func(i, j int) bool {
		return history[i].Date > history[j].Date
	})

	return history, nil
}

// getHistoryEntryFromFile reads a single CSV file and extracts the total hours worked
func getHistoryEntryFromFile(filename, date string) (HistoryEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return HistoryEntry{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = 3

	records, err := reader.ReadAll()
	if err != nil {
		return HistoryEntry{}, err
	}

	if len(records) == 0 {
		return HistoryEntry{Date: date, Hours: 0}, nil
	}

	// Get the last record to find the total time worked
	lastRecord := records[len(records)-1]
	timeStr := lastRecord[1] // Format: HH:MM:SS

	// Parse the time string
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return HistoryEntry{}, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.Atoi(parts[2])

	// Convert to decimal hours
	totalHours := float64(hours) + float64(minutes)/60.0 + float64(seconds)/3600.0

	// Round to 1 decimal place
	totalHours = math.Round(totalHours*10) / 10

	return HistoryEntry{Date: date, Hours: totalHours}, nil
}
