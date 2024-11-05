package readers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func HTTPEvents(ctx context.Context) <-chan []byte {
	// Create a channel to send the request bodies
	bodyChannel := make(chan []byte)

	// Define the HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Send the body to the channel
		bodyChannel <- body

		// Respond to the client
		fmt.Fprintf(w, "Request received")
	})

	go func() {
		defer func() {
			close(bodyChannel)
		}()

		go func() {
			// Start the HTTP server
			log.Println("Starting server on :8080")
			if err := http.ListenAndServe(":8080", nil); err != nil {
				log.Fatalf("Server failed: %s\n", err)
			}
		}()

		for {
			time.Sleep(1 * time.Second)
			select {
			case <-ctx.Done():
				// kill the server
				return
			default:
				//
			}
		}
	}()

	return bodyChannel
}
