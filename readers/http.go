package readers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func HTTPEvents() <-chan []byte {
	// Create a channel to send the request bodies
	bodyChannel := make(chan []byte)

	// Define the HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body, err := ioutil.ReadAll(r.Body)
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
		// Start the HTTP server
		log.Println("Starting server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Server failed: %s\n", err)
		}
	}()

	return bodyChannel
}
