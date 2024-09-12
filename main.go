package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

var overflowChannel = make(chan string, 10) // Buffered channel with capacity of 10

func overflowHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	select {
	case overflowChannel <- string(body):
		fmt.Fprintf(w, "Request added to the channel!")
	default:
		fmt.Fprintf(w, "Channel is full, request not added.")
		if err := OverflowToQueue(string(body)); err != nil {
			http.Error(w, "Failed to write to overflow queue", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func processReqeust(ctx context.Context) {
	for {
		select {
		case msg := <-overflowChannel:
			fmt.Println("Processing:", msg)
		case <-ctx.Done():
			fmt.Println("Context closed, stopping processing.")
			return
		}
	}
}

func OverflowToQueue(body string) error {
	// todo write to external queue
	return nil
}

func main() {
	go processReqeust(context.Background()) // Start the goroutine to process the channel

	http.HandleFunc("/overflow", overflowHandler)
	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
