package main

import (
	"flag"
	"github.com/zeebo/sbloom"
	"hash/fnv"
	"log"
	"os"
)

const probability = 10

var (
	filters        map[string]sbloom.Filter
	filterFilename string
	searchMode     bool
)

func init() {
	flag.StringVar(&filterFilename, "f", "/tmp/BloomFilterStore", "Filename to read and store the filter state")
	flag.BoolVar(&searchMode, "s", false, "Whether arguments are search terms or files to store")
}

func storeFile(f *os.File, err error) {
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	filters[f.Name()] = *sbloom.NewFilter(fnv.New64(), 10)

	b := make([]byte, 1)
	word := make([]byte, 64)
	for {
		if _, err := f.Read(b); err != nil {
			break
		}
		log.Printf("Letter/word: %s/%s", b, word)
	}
}

func main() {
	flag.Parse()
	log.Printf("Filter File: %s", filterFilename)
	log.Printf("Search Mode: %v", searchMode)
	log.Printf("Args:        %+v", flag.Args())

	filters = make(map[string]sbloom.Filter)
	//sbloom.NewFilter(fnv.New64(), 10)

	switch {
	case searchMode:
		log.Println("Unimplemented")
	case len(flag.Args()) > 0:
		for _, f := range flag.Args() {
			storeFile(os.Open(f))
		}
	}
}
