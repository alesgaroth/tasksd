package main

import (
	"log"
	"net/http"

	"github.com/alesgaroth/tasksd/internal/web"
)

func main() {
	handler := web.NewServer()
	addr := ":8080"
	log.Printf("Starting tasksd on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
