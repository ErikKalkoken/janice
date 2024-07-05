package main

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
)

const (
	maxKeys   = 11
	maxLevels = 6
	fileName  = "test.json"
)

var n = 0

func main() {
	fmt.Printf("Generating JSON file with %d keys per object and %d levels...\n", maxKeys, maxLevels)
	obj := makeObj(0)
	fmt.Println("Marshalling into JSON...")
	b, err := json.Marshal(obj)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Writing file: %s ...", fileName)
	if err := os.WriteFile(fileName, b, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated JSON file with %d elements.\n", n)
}

func makeObj(level int) map[string]any {
	o := make(map[string]any)
	for i := range maxKeys {
		if level == 0 {
			fmt.Printf("Generating %d / %d\r", i+1, maxKeys)
		}
		k := randomBase16String(10)
		if level < maxLevels {
			o[k] = makeObj(level + 1)
		} else {
			o[k] = rand.Intn(10_000)
		}
		n++
	}
	return o
}

func randomBase16String(l int) string {
	buff := make([]byte, int(math.Ceil(float64(l)/2)))
	cryptorand.Read(buff)
	str := hex.EncodeToString(buff)
	return str[:l] // strip 1 extra character we get from odd length results
}
