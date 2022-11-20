package main

/*
Output:
worker 3 started job 1
worker 1 started job 2
worker 2 started job 3
worker 2 finished job 3
worker 1 finished job 2
worker 3 finished job 1
*/

import (
	"fmt"
	"time"
)

// Worker receives work on jobs channel and send results on results channel
func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Println("worker", id, "started job", j)
		time.Sleep(time.Second) // Simulate work
		fmt.Println("worker", id, "finished job", j)
		results <- j * 2 // Send result
	}
}

// Execute 3 jobs on 3 workers
// Fast because total work is split across workers concurrently
func pool() {
	// Create channels to send and collect results
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	// Start N concurrent workers (initially blocked due to no jobs)
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	// Send N jobs and close channel to indicate all work has been sent
	for j := 1; j <= 3; j++ {
		jobs <- j // Send work on jobs channel
	}
	close(jobs)

	// Receive N results on results channel
	for r := 1; r <= 3; r++ {
		<-results // Collect results of work
	}
}

func main() {
	pool()
	selectExample()
}

// Select: wait on multiple channel operations
// <- c1 : receive message from channel c1
// c1 <- data : send data on channel c1
// Output: received one, received two (vice versa)
func selectExample() {

	c1 := make(chan string) // Unbuffered channel blocks automatically to synchronize between goroutines
	c2 := make(chan string)

	// Each channel will receive value after simulated work (e.g. blocking RPC operations executing in concurrent goroutines )
	go func() {
		time.Sleep(time.Second * 1) // Simulate work
		c1 <- "one"
	}()
	go func() {
		time.Sleep(time.Second * 1) // Simulate work
		c2 <- "two"
	}()

	// Use select to await both values simultaneously on different channels
	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-c1:
			fmt.Println("received", msg1)
		case msg2 := <-c2:
			fmt.Println("received", msg2)
		}
	}
}
