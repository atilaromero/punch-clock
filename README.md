# Punch Clock

A simple web-based punch clock application built with Go.

## Features

- Start and pause work tracking
- Displays current worked time (hours, minutes, seconds) in the page tab
- Automatically saves data to a CSV file with the current date (YYYY-MM-DD.csv)
- Loads existing data for the current day on startup
- Tracks total worked time across multiple sessions

## CSV File Format

The application saves data in a comma-separated values (CSV) file with the following fields:
1. Action (start/pause)
2. Total worked time (HH:MM:SS)
3. ISO timestamp (RFC3339 format)

## Running the Application

```bash
# Navigate to the project directory
cd punch-clock

# Run the application
go run main.go
```

The application will start a web server on http://localhost:8080.

## Requirements

- Go 1.16 or higher

## Usage

1. Open http://localhost:8080 in your web browser
2. Click "Start" to begin tracking time
3. Click "Pause" to stop tracking time
4. The current worked time is displayed in the browser tab and on the page
5. Data is automatically saved to a CSV file named with the current date
