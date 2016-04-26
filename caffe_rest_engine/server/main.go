package main

import (
	"flag"
	"fmt"
	"net/http"
)

var (
	nworkers = flag.Int("n", 4, "Enter the number of workers wanted.")
)

func main() {

	flag.Parse()

	fmt.Println("Starting the dispatcher!")
	fmt.Println("nworker %d", *nworkers)
	dis := NewDispatcher("placeholder", *nworkers)
	dis.StartDispatcher()

	fmt.Println("Registering HTTP Function")
	http.HandleFunc("/classify", JobHandler)

	fmt.Println("HTTP Server listening on 127.0.0.1:8000")
	err := http.ListenAndServe("localhost:8000", nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}
