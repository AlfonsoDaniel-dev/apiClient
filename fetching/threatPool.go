package fetching

import (
	"errors"
	"net/http"
	"time"
)

// FetchingAttempt represents a single HTTP request attempt with its associated data
type FetchingAttempt struct {
	id                 int    // Unique identifier for this attempt
	url                string // Target URL for the HTTP request
	method             string // HTTP method (GET, POST, etc.)
	Data               string // Response data from the HTTP request
	ErrorDuringRequest error  // Error that occurred during the request
	StatusCode         int    // HTTP status code from the response
}

// NewfetchingAttempt creates a new FetchingAttempt with the given parameters
func NewfetchingAttempt(Url, Method string, id int) *FetchingAttempt {
	return &FetchingAttempt{
		id:     id,
		url:    Url,
		method: Method,
	}
}

// FetchingThreadPool manages a pool of workers to handle HTTP requests concurrently
type FetchingThreadPool struct {
	maxWorker      uint                  // Maximum number of worker goroutines
	Attempts       []*FetchingAttempt    // List of all HTTP requests to be processed
	jobsChan       chan *FetchingAttempt // Channel for distributing jobs to workers
	resolveChannel chan FetchingAttempt  // Channel for collecting completed results
	workers        []*fetchingWorker     // Slice of worker instances
	client         *http.Client          // HTTP client for making requests
}

// NewFetchingThreadPool creates a new thread pool with the specified number of workers
func NewFetchingThreadPool(attempts []*FetchingAttempt, maxWorkers uint) (*FetchingThreadPool, error) {

	// Validate input parameters
	if len(attempts) <= 0 {
		return nil, errors.New("wrong parameters while creating new fetching threadpool")
	}

	// Create the thread pool instance
	threadPool := &FetchingThreadPool{
		maxWorker:      maxWorkers,
		Attempts:       attempts,
		jobsChan:       make(chan *FetchingAttempt, len(attempts)), // Buffered channel for job distribution
		resolveChannel: make(chan FetchingAttempt, len(attempts)),  // Buffered channel for result collection
		workers:        make([]*fetchingWorker, maxWorkers),        // Initialize worker slice
		client: &http.Client{
			Timeout: 5 * time.Second, // Set timeout for HTTP requests
		},
	}

	// Create and start all worker goroutines
	for i := 0; i < int(maxWorkers); i++ {
		worker := NewFetchingWorker(uint(i), threadPool)

		threadPool.workers[i] = worker

		go worker.work() // Start worker in background
	}

	return threadPool, nil
}

// listenResolveChan listens to the resolve channel and collects all completed results
func (tp *FetchingThreadPool) listenResolveChan() []FetchingAttempt {

	var resolves []FetchingAttempt // Slice to store all completed results
	var count int                  // Counter for completed requests

	// Listen for completed requests until all are received
	for {

		resolve := <-tp.resolveChannel // Wait for next completed request

		resolves = append(resolves, resolve) // Add to results slice

		count++ // Increment counter

		// Check if all requests have been processed
		if count == len(tp.Attempts) {
			return resolves
		}

	}
}

// FetchData distributes all HTTP requests to workers and waits for completion
func (tp *FetchingThreadPool) FetchData() []FetchingAttempt {

	// Send all HTTP requests to the jobs channel for worker processing
	for _, attempt := range tp.Attempts {

		tp.jobsChan <- attempt // Send job to available worker

	}

	// Wait for all responses to be collected
	resolves := tp.listenResolveChan()

	// Close the resolve channel to signal completion
	close(tp.resolveChannel)

	// Send termination signals to all workers
	for _, w := range tp.workers {

		w.killchan <- struct{}{} // Send kill signal to worker

	}

	return resolves
}
