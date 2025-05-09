package main

import (
	"log"
	"net/http"
)

const port string = "42069"

func main() {
	mux := http.NewServeMux()
	filepathRoot := http.Dir(".")

	mux.Handle("/assets/", http.FileServer(filepathRoot))

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	server.ListenAndServe()
}
