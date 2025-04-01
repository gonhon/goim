package main

import (
	"log"
	"net/http"
)

func main() {
	// Simple static webserver:
	log.Print("Lisen port:[1999]")
	log.Fatal(http.ListenAndServe(":1999", http.FileServer(http.Dir("./"))))
}
