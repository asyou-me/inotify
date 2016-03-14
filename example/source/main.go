package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("/tmp")))
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("http server is run with :8080")
	}()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
