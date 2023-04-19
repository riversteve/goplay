package main

import (
	"fmt"
	"net/http"
	"time"
)

type LogMessage struct {
	timestamp time.Time
	message   string
}

var logChannel = make(chan LogMessage)

func generateLogs() {
	for {
		time.Sleep(1 * time.Second)
		logMsg := LogMessage{
			timestamp: time.Now(),
			message:   fmt.Sprintf("Log message at %s", time.Now().Format(time.RFC3339)),
		}
		logChannel <- logMsg
	}
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	for logMsg := range logChannel {
		fmt.Fprintf(w, "data: %s\n\n", logMsg.message)
		flusher.Flush()
	}
}

func main() {
	go generateLogs()

	http.HandleFunc("/sse", handleSSE)
	// Serve the index.html file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":8080", nil)
}
