package main

import (
	"log"

	"github.com/felip/api-fidelidade/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("bootstrap application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("run application: %v", err)
	}
}
