package main

import (
	"encoding/json"
	"log"
	"os"

	"example/jsonviewer/internal/ui"
)

func main() {
	path := ".temp/meta.json"
	dat, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read file %s: %s", path, err)
	}
	var data any
	if err := json.Unmarshal(dat, &data); err != nil {
		log.Fatalf("failed to unmarshal JSON: %s", err)
	}
	log.Printf("Read and unmarshaled JSON file %s", path)
	u := ui.NewUI()
	if err := u.SetData(data); err != nil {
		log.Fatal(err)
	}
	u.ShowAndRun()
}
