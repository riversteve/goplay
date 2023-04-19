package main

import (
	"fmt"
	"net/http"
	"time"
)

type LogMessage struct {
	timestamp  time.Time
	level      string
	sourceFunc string
	message    string
}

var logChannel = make(chan LogMessage)

func formatLogMessage(level, sourceFunc, message string) LogMessage {
	return LogMessage{
		timestamp:  time.Now(),
		level:      level,
		sourceFunc: sourceFunc,
		message:    message,
	}
}

func generateLogs() {
	for {
		time.Sleep(1 * time.Second)
		logMsg := formatLogMessage("INFO", "generateLogs", fmt.Sprintf("Log message at %s", time.Now().Format(time.RFC3339)))
		logChannel <- logMsg
	}
}

func anotherFunction() {
	for {
		time.Sleep(2 * time.Second)
		logMsg := formatLogMessage("DEBUG", "anotherFunction", fmt.Sprintf("Another function log at %s", time.Now().Format(time.RFC3339)))
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
		logData := fmt.Sprintf("[%s] [%s] [%s] %s", logMsg.timestamp.Format(time.RFC3339), logMsg.level, logMsg.sourceFunc, logMsg.message)
		fmt.Fprintf(w, "data: %s\n\n", logData)
		flusher.Flush()
	}
}

func main() {
	go generateLogs()
	go anotherFunction()

	http.HandleFunc("/sse", handleSSE)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":8080", nil)
}
