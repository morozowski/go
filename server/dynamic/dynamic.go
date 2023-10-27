package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

type page struct {
	path        string
	description string
	function    func(http.ResponseWriter, *http.Request)
}

var pages = []page{}

func setupPage(path, description string, handler func(http.ResponseWriter, *http.Request)) {
	pages = append(pages, page{path, description, handler})
	http.HandleFunc(path, handler)
}

func appendBottom(w http.ResponseWriter, requestURI string) {
	for _, p := range pages {
		if p.path != requestURI {
			fmt.Fprintf(w, "[<a href=\"%s\">%s</a>]\n", p.path, p.description)
		}
	}
}

func numCPU(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<title>CPU</title>")
	fmt.Fprintf(w, "<h1>CPUs: %d</h1>\n", runtime.NumCPU())
	appendBottom(w, r.RequestURI)
}

func now(w http.ResponseWriter, r *http.Request) {
	s := time.Now().Format("02/01/2006 15:04:05 MST")
	fmt.Fprintln(w, "<title>Local Time</title>")
	fmt.Fprintln(w, "<meta http-equiv=\"refresh\" content=\"1\">")
	fmt.Fprintf(w, "<h1>Local time: %s</h1>\n", s)
	appendBottom(w, r.RequestURI)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<title>Hello World</title>")
	fmt.Fprintln(w, "<h1>Hello World!</h1>")
	appendBottom(w, r.RequestURI)
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatal("Missing parameter: <port>")
	}
	port := args[0]

	setupPage("/", "Home", home)
	setupPage("/time", "Local Time", now)
	setupPage("/cpu", "CPU", numCPU)

	log.Println("Listening to port " + port + "...")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
