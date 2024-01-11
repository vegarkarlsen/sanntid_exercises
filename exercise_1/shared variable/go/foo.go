// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
    "time"
)

var i = 0

func incrementing() {
    //TODO: increment i 1000000 times
    for a := 0; a < 1000000; a++{
        Println("dec");
        i++; 
    }

}

func decrementing() {
    //TODO: decrement i 1000000 times
    for a := 0; a < 1000000; a++ {
        Println("dec");
        i--;
    }
}

func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    // Maximum threads that can run, if we change to on then one routine will run first and then the other afterwords. instead of "simoultanizaly"
    runtime.GOMAXPROCS(2)    
	
    // TODO: Spawn both functions as goroutines
    go incrementing();
    go decrementing();

	
    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    time.Sleep(500*time.Millisecond)
    Println("The magic number is:", i)
}