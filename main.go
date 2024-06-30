package main

import (
	"log"

	"github.com/ErikKalkoken/jsonviewer/internal/ui"
)

func main() {
	u, err := ui.NewUI()
	if err != nil {
		log.Fatalf("Failed to initialize application: %s", err)
	}
	u.ShowAndRun()
}
