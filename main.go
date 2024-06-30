package main

import (
	"example/jsonviewer/internal/ui"
	"log"
)

func main() {
	u, err := ui.NewUI()
	if err != nil {
		log.Fatalf("Failed to initialize application: %s", err)
	}
	u.ShowAndRun()
}
