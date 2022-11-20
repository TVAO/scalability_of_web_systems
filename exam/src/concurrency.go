package main

import "fmt"

// Concurrency: app is making progress on more than one task at the same time (concurrently)
// Parallelism: application splits its tasks up into smaller subtasks (concurrent design) and runs them in parallel on multiple CPUs
// Link: http://tutorials.jenkov.com/java-concurrency/concurrency-vs-parallelism.html
// Concurrency: https://www.golang-book.com/books/intro/10

// Go supports concurrency using goroutines and channels
// A goroutine is a lightweight thread managed by Go runtime
func f(n int) {
	for i := 0; i < 10; i++ {
		fmt.Println(n, ":", i)
	}
}

// Two routines: implicit main function itself + go f(0)
func routineExample() {
	go f(0)
	var input string
	fmt.Scanln(&input)
}

// Channels provide a way for two goroutines to communicate with one another and synchronize their execution
// Concurrency by message passing instead of using locks on shared mutable state

// A channel is a typed conduit (pipeline) through which you can send and receive values via <- operator
// ch <- v // Send v to channel ch
// v := <-ch // Receive from ch and assign value to v
// (data flows in direction of arrow)
// Default: sends and receives block until other side is ready -> allows goroutines to synchronize without explicit locks or condition variables

// Example code: sums numbers in slice, distributing work between two goroutines
// Once both goroutines complete, it calculates final result
func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // send sum to c (send blocks by default allowing goroutine to sync w/o lock)
}

func channelExample() {
	s := []int{7, 2, 8, -9, 4, 0}
	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // receive from c blocks
	fmt.Println(x, y, x+y)
}

// Buffered channel: sends to buffered channel block only when buffer is full
// Receives block when buffer is empty
func bufferedChannel() {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	//ch <- 3 // deadlock due to overfilling buffer
	fmt.Println(<-ch)
	fmt.Println(<-ch)
}

// Close: close channel to indicate no more values will be sent
// Check: v, ok := <- ch // ok is false if closed (there are no more values to receive)
// NB: ONLY CLOSE CHANNEL ON SENDER AND DO SO IF RECEIVER MUST BE TOLD NO MORE VALUES COMING
func fibonacci(n int, c chan int) {
	x, y := 0, 1
	for i := 0; i < n; i++ {
		c <- x
		x, y = y, x+y
	}
	close(c)
}
// Prints 0 1 1 2 3 5 8 13 21 34
func closeExample() {
	func main() {
		c := make(chan int, 10)
		go fibonacci(cap(c), c)
		for i := range c { // Receives value from sender channel repeatedly until it is closed 
			fmt.Println(i)
		}
	}
}

// Select: lets a goroutine wait on multiple communication operations 
// Blocks until one of its cases can run, then executes that case (random chosen if multiple ready)
func fibonacciSelect(c, quit chan int) {
	x, y := 0, 1
	for {
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("quit")
			return
		}

	}
}
func selectExample() {
	c := make(chan int)
	quit := make(chan int)
	go func() { // Anonymous go function 
		for i := 0; i < 10; i++ {
			fmt.Println(<-c) // Receive results from channel c 
		}
		quit <- 0 // Send value to channel quit to stop program 
	}()
}

// Default: default case in select is run if no case is ready
// Use a default case to try to send or receive without blocking
/*

select {
case i := <-c:
    // use i
default:
    // receiving from c would block
}

*/

func defaultExample() {
	tick := time.Tick(100 * time.Millisecond)
	boom := time.After(500 * time.Millisecond)
	for {
		select {
		case <-tick:
			fmt.Println("tick.")
		case <-boom:
			fmt.Println("BOOM!")
			return
		default:
			fmt.Println("    .")
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// Mutual exclusion using mutex if we want goroutines to access shared mutable variable w/o race conditions 
func mutexExample() {
	sharedCount := 0
	for i := 0; i < 100; i++ {
		lock.Lock()
		go func() {
			lock := sync.Mutex
			sharedCount += 1
			lock.Unlock()
		}()
	}
}



