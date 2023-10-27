package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatal("Missing parameter: <port>")
	}
	port := args[0]

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	log.Println("Listening to port " + port + "...")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
