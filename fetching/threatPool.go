package fetching

import (
	"errors"
	"net/http"
	"reflect"
	"time"
)

func isSchemaOk(val any) bool {

	reflectValue := reflect.ValueOf(val)

	if reflectValue.Kind() != reflect.Ptr {
		return false
	} else if reflectValue.Elem().Kind() != reflect.Struct {
		return false
	}

	return true
}

// FetchingAttempt represents a single HTTP request attempt with its associated data
type FetchingAttempt struct {
	id             int    // Unique identifier for this attempt
	url            string // Target URL for the HTTP request
	method         string // HTTP method (GET, POST, etc.)
	ResponseSchema any    // The kind of value you want to get
}

// NewfetchingAttempt creates a new FetchingAttempt with the given parameters
func NewfetchingAttempt(Url, Method string, id int, responseKind any) (*FetchingAttempt, error) {

	if !isSchemaOk(responseKind) {
		return nil, errors.New("the result format you want must be a pointer")
	}

	return &FetchingAttempt{
		id:     id,
		url:    Url,
		method: Method,
	}, nil
}

type FetchingResult struct {
	id                 int
	ErrorDuringRequest error // Error that occurred during the request
	StatusCode         int   // HTTP status code from the response
	Data               any
}

func newFetchingresult(id int, data any) *FetchingResult {
	return &FetchingResult{
		id:   id,
		Data: data,
	}
}

// FetchingThreadPool manages a pool of workers to handle HTTP requests concurrently
type FetchingThreadPool struct {
	maxWorker      uint // Maximum number of worker goroutines
	attemptsQueue  chan *FetchingAttempt
	jobsChan       chan *FetchingAttempt // Channel for distributing jobs to workers
	resolveChannel chan FetchingResult   // Channel for collecting completed results
	Results        map[int]*FetchingResult
	workers        []*fetchingWorker // Slice of worker instances
	client         http.Client       // HTTP client for making requests
}

// NewFetchingThreadPool creates a new thread pool with the specified number of workers
func NewFetchingThreadPool(maxWorkers uint) (*FetchingThreadPool, error) {

	// Validate input parameters
	if maxWorkers < 1 {
		return nil, errors.New("wrong parameters while creating new fetching threadpool")
	}

	// Create the thread pool instance
	threadPool := &FetchingThreadPool{
		maxWorker:      maxWorkers,
		Attempts:       attempts,
		jobsChan:       make(chan *FetchingAttempt), // Buffered channel for job distribution
		resolveChannel: make(chan FetchingResult),   // Buffered channel for result collection
		client: http.Client{
			Timeout: 5 * time.Second, // Set timeout for HTTP requests
		},
	}

	// Create and start all worker goroutines
	for i := 0; i < int(maxWorkers); i++ {
		worker := newFetchingWorker(uint(i), threadPool)

		threadPool.workers[i] = worker

		go worker.work() // Start worker in the background
	}

	return threadPool, nil
}

func (tp *FetchingThreadPool) Start() {

}

func (tp *FetchingThreadPool) Use() {

}

// listenResolveChan listens to the resolve channel and collects all completed results
func (tp *FetchingThreadPool) listenResolveChan() {

	var resolves []FetchingAttempt // Slice to store all completed results
	var count int                  // Counter for completed requests

	// Listen for completed requests until all are received
	for {

		resolve := <-tp.resolveChannel // Wait for the next completed request

		resolves = append(resolves, resolve) // Add to the results slice

		count++ // Increment counter

		// Check if all requests have been processed

	}
}

// FetchData distributes all HTTP requests to workers and waits for completion
func (tp *FetchingThreadPool) interact() []FetchingAttempt {

	// Send all HTTP requests to the job channel for worker processing
	for _, attempt := range tp.Attempts {

		tp.jobsChan <- attempt // Send the job to an available worker

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

func (tp *FetchingThreadPool) startWorkers() {

	var workers = make([]*fetchingWorker, tp.maxWorker)

	for i := range tp.workers {

		worker := newFetchingWorker(uint(i), tp)

		workers = append(tp.workers, worker)

		go worker.work()
	}

	tp.workers = workers

}
