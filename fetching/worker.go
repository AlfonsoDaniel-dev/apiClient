package fetching

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// fetchingWorker represents a single worker goroutine that processes HTTP requests
type fetchingWorker struct {
	Id         uint                // Unique identifier for this worker
	ThreadPool *FetchingThreadPool // Reference to the parent thread pool
	killchan   chan struct{}       // Channel for receiving termination signals
}

// NewFetchingWorker creates a new worker instance with the given ID and thread pool reference
func NewFetchingWorker(id uint, tp *FetchingThreadPool) *fetchingWorker {
	newWorker := &fetchingWorker{
		Id:         id,
		ThreadPool: tp,
		killchan:   make(chan struct{}), // Initialize kill channel
	}

	return newWorker
}

// work is the main worker loop that processes HTTP requests until terminated
func (fw *fetchingWorker) work() {

	// Main worker loop - continues until kill signal is received
	for {

		select {

		case attempt := <-fw.ThreadPool.jobsChan:
			// Process a new HTTP request job

			var res *http.Response
			res, attempt.ErrorDuringRequest = fw.getData(attempt.url) // Make HTTP request

			attempt.StatusCode = res.StatusCode // Set response status code

			attempt.Data, attempt.ErrorDuringRequest = fw.readData(res) // Read response body

			fw.ThreadPool.resolveChannel <- *attempt // Send completed result back

		case <-fw.killchan:
			// Received termination signal - clean up and exit
			close(fw.killchan) // Close kill channel
			return             // Exit worker loop
		}

	}

}

// readData reads the response body from an HTTP response and returns it as a string
func (fw *fetchingWorker) readData(res *http.Response) (string, error) {

	// Check if response is nil to avoid panic
	if res == nil {
		return "", errors.New("response is nil")
	}

	var err error

	// Read all data from response body
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil // Convert bytes to string
}

// getData performs an HTTP GET request to the specified URL and returns the response
func (fw *fetchingWorker) getData(url string) (*http.Response, error) {

	// Use the thread pool's HTTP client to make the request
	res, err := fw.ThreadPool.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error while fetching Data. Err: %v", err)
	}

	return res, err
}
