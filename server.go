package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	// run multiple instances of this server on different ports
	// go run server.go 8081
	// go run server.go 8082
	// go run server.go 8083

	port := "8080" // Default port
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		delayStr := r.URL.Query().Get("delay")
		delayMs := 5 // add some delay as there is actually no backend
		if delayStr != "" {
			if d, err := strconv.Atoi(delayStr); err == nil {
				delayMs = d
			}
		}

		log.Printf("Backend Server %s: Received request. Simulating %dms delay.", port, delayMs)
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		fmt.Fprintf(w, "Hello from Backend Server on port %s! Processed in %dms.\n", port, delayMs)
	})

	log.Printf("Starting Backend Server on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Backend server failed: %v", err)
	}
}
