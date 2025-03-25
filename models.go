package main

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
