package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	u := newUI()
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
	if err := u.setData(data); err != nil {
		log.Fatal(err)
	}
	u.showAndRun()
}
