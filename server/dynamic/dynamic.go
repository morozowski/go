package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

func numCPU(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<title>CPU</title>")
	fmt.Fprintf(w, "<h1>CPUs: %d</h1>\n", runtime.NumCPU())
}

func now(w http.ResponseWriter, r *http.Request) {
	s := time.Now().Format("02/01/2006 03:04:05")
	fmt.Fprintln(w, "<title>Local Time</title>")
	fmt.Fprintln(w, "<meta http-equiv=\"refresh\" content=\"1\">")
	fmt.Fprintf(w, "<h1>Local time: %s</h1>\n", s)
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatal("Missing parameter: <port>")
	}
	port := args[0]

	http.HandleFunc("/cpu", numCPU)
	http.HandleFunc("/time", now)

	log.Println("Listening to port " + port + "...")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
