package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var process = make(chan string, 10) // Buffered channel with capacity of 10

func processHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	select {
	case process <- string(body):
		log.Println("request added to the channel!")
	default:
		log.Println("channel is full, request not added.")
		if err := overflowToQueue(string(body)); err != nil {
			http.Error(w, "failed to write to overflow queue", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func processReqeust(ctx context.Context) {
	for {
		select {
		case msg := <-process:
			log.Println("Processing:", msg)
			// fake work!
			time.Sleep(5 * time.Second)
		case <-ctx.Done():
			log.Println("context closed, stopping processing.")
			return
		}
	}
}

func overflowToQueue(body string) error {
	log.Println("overflowing, sending body to queue", body)
	// todo write to external queue
	return nil
}

func main() {
	go processReqeust(context.Background()) // Start the goroutine to process the channel

	http.HandleFunc("/process", processHandler)
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println("Server failed to start:", err)
		os.Exit(1)
	}
}
