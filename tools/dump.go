/*
Package tools implements basic tools such as stack on signal dumping 
*/
package tools

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// Dump dumps the application stack it has been called by when a quit Signal is received
// by stroucki: https://gist.github.com/stroucki/0b60ebce40139d1edb93bf20006721bf
func Dump() {
	c := make(chan os.Signal, 1)
	// handle signal: SIGUSR1 10 0xa,  see 'kill -l'
	// kill -10 <pid>
	// kill -SIGUSR1 <pid>
	// kill -s SIGUSR1 <pid>
	signal.Notify(c, syscall.Signal(0xa))
	go func() {
		for range c {
			Stacks()
		}
	}()
}

// Stacks dumps full stacks in current process
func Stacks() {
	buf := make([]byte, 1<<24)
	buf = buf[:runtime.Stack(buf, true)]
	fmt.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
}
