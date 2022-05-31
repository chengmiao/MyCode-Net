package main

import (
	"fmt"
	"os"
	"syscall"
	"os/signal"
)

// go echo server
func main() {






	// test := []string{"567", "567", "567", "123", "123"}

	// tmp := append([]string{}, test...)

	// fmt.Println("slice is : ", tmp)
    
	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGINT)

	fmt.Println("Press Ctrl+C to exit")
	<-done
	fmt.Println("\nBye")

}