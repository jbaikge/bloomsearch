package main

import (
	"bytes"
	"flag"
	"github.com/zeebo/sbloom"
	"hash/fnv"
	"log"
	"os"
)

const (
	probability = 10
)

var (
	accept         = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	filters        map[string]sbloom.Filter
	filterFilename string
	searchMode     bool
	wordbreaks     = []byte("\n\r ")
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

	filter := sbloom.NewFilter(fnv.New64(), 10)

	b := make([]byte, 1)
	word := make([]byte, 64)
	for {
		if _, err := f.Read(b); err != nil {
			break
		}
		switch {
		// Store a word on a break
		case bytes.Contains(wordbreaks, b):
			filter.Add(word)
			filter.Add(bytes.ToLower(word))
			word = word[:0]
		case bytes.Contains(accept, b):
			word = append(word, b[0])
		}
		log.Printf("Letter/word: %s/%s", b, word)
	}
	filters[f.Name()] = *filter
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
