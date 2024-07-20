// This is a tool for generating large JSON files by multiplying an existing json document
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	factorDefault = 3
)

var factorFlag = flag.Int("f", factorDefault, "how many times to multiply the contents of a JSON doc")

func main() {
	flag.Parse()
	input := flag.Arg(0)
	if input == "" {
		log.Fatal("Need to specify an input file")
	}
	path, err := filepath.Abs(input)
	if err != nil {
		log.Fatalf("Error reading %s", input)
	}
	fmt.Printf("Reading %s...\n", path)
	data, err := os.ReadFile(input)
	if err != nil {
		log.Fatalf("Error reading %s", path)
	}
	fmt.Println("Un-marshalling JSON...")
	var source any
	if err := json.Unmarshal(data, &source); err != nil {
		log.Fatalf("Error un-marshalling %s", path)
	}
	data = nil
	fmt.Println("Multiplying data...")
	target := make(map[string]any)
	for i := range *factorFlag {
		k := fmt.Sprintf("clone_%03d", i)
		target[k] = source
	}
	filename := "out.json"
	fmt.Printf("Writing file: %s ...", filename)

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	if err := enc.Encode(target); err != nil {
		log.Fatal(err)
	}
	fmt.Println("DONE")
}
