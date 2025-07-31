package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlfonsoDaniel-dev/apiClient/fetching"
)

// main is the entry point of the application
// It demonstrates the usage of the FetchingThreadPool to make concurrent HTTP requests
func main() {

	start := time.Now() // time where the program started

	// Define the number of concurrent HTTP requests to make
	numofcalls := 4

	// Create a slice to hold all HTTP request attempts
	var attempts []*fetching.FetchingAttempt

	// Create HTTP request attempts for the specified number of calls
	for i := 0; i < numofcalls; i++ {
		// Create a new attempt to fetch data from the JSONPlaceholder API
		attempt := fetching.NewfetchingAttempt("https://rickandmortyapi.com/api/character/93", http.MethodGet, i)

		attempts = append(attempts, attempt)
	}

	// Create a new thread pool with the attempts and specify maximum number of workers
	threadPool, err := fetching.NewFetchingThreadPool(attempts, 2)
	if err != nil {
		log.Fatal("error while creating threadPool")
	}

	// Execute all HTTP requests concurrently and wait for all responses
	responses := threadPool.FetchData()

	// Process and display the results of all HTTP requests
	for i, response := range responses {

		// Also print just the data for easier reading
		fmt.Printf("\n Petition Number: %d . Status: %d . Data: %v .", i, response.StatusCode, response.Data)
	}

	// prints num of calls and the time used to query them
	fmt.Printf("\n Number of querys: %d .Execution Time: %v .", numofcalls, time.Since(start).Seconds())

}
