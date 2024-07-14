package main

import (
	"bufio"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	keysDefault   = 8
	levelsDefault = 3
	fileName      = "test.json"
)

var n = 0

var keysFlag = flag.Int("k", keysDefault, "number of keys generated per object")
var levelsFlag = flag.Int("l", levelsDefault, "number of generated levels")

func main() {
	flag.Parse()
	fmt.Printf("You have selected %d keys and %d levels.\n", *keysFlag, *levelsFlag)
	fmt.Println("This can take a while. Are you sure you want to continue (Y/n)?")
	consoleReader := bufio.NewReaderSize(os.Stdin, 1)
	input, _ := consoleReader.ReadByte()
	if input == 'n' {
		return
	}
	fmt.Println("Generating JSON file...")
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
	p := message.NewPrinter(language.English)
	p.Printf("Generated JSON file with %d elements.\n", n+1)
}

func makeObj(level int) map[string]any {
	o := make(map[string]any)
	for i := range *keysFlag {
		if level == 0 {
			fmt.Printf("Generating %d / %d\r", i+1, *keysFlag)
		}
		k := randomBase16String(10)
		if level < *levelsFlag {
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
