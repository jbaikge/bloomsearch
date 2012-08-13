package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
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
	lowerMap       [256]byte
)

func init() {
	flag.StringVar(&filterFilename, "f", "/tmp/BloomFilterStore", "Filename to read and store the filter state")
	flag.BoolVar(&searchMode, "s", false, "Whether arguments are search terms or files to store")
	for i := byte(1); i > 0; i++ {
		if 'A' <= i && i <= 'Z' {
			lowerMap[i] = 'a' + (i - 'A')
		} else {
			lowerMap[i] = i
		}
	}
}

func restoreFilters(name string) (err error) {
	log.Printf("Loading previous gob from %s", name)
	// Attempt to read in the previous filters
	f, err := os.Open(name)
	if err != nil {
		return
	}
	defer f.Close()
	err = gob.NewDecoder(f).Decode(&filters)
	if err != nil {
		return
	}
	log.Printf("Done")
	return
}

func saveFilters(name string) (err error) {
	log.Printf("Saving gob to %s", name)
	f, err := os.Create(name)
	if err != nil {
		return
	}
	defer f.Close()
	err = gob.NewEncoder(f).Encode(&filters)
	if err != nil {
		return
	}
	log.Println("Done")
	return
}

func search(words ...string) (files []string) {
	var found bool
	for f, filter := range filters {
		found = true
		for _, w := range words {
			found = found && filter.Lookup([]byte(w))
		}
		if found {
			files = append(files, f)
		}
	}
	return
}

func storeFile(f *os.File, err error) {
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	filter := sbloom.NewFilter(fnv.New64(), 10)

	b := make([]byte, 1)
	word := make([]byte, 0, 64)
	for {
		if _, err := f.Read(b); err != nil {
			break
		}
		switch {
		// Store a word on a break
		case bytes.Contains(wordbreaks, b):
			filter.Add(word)
			toLower(word)
			filter.Add(word)
			word = word[:0]
		case bytes.Contains(accept, b):
			word = append(word, b[0])
		}
	}
	filters[f.Name()] = *filter
}

func toLower(p []byte) {
	for i, v := range p {
		p[i] = lowerMap[v]
	}
}

func main() {
	flag.Parse()
	log.Printf("Filter File: %s", filterFilename)
	log.Printf("Search Mode: %v", searchMode)
	log.Printf("Args Length: %+v", len(flag.Args()))

	filters = make(map[string]sbloom.Filter)

	if err := restoreFilters(filterFilename); err != nil {
		log.Printf("Could not restore filters: %s; Continuing with blank filter list", err)
	}

	switch {
	case searchMode:
		log.Println("Searching...")
		for i, f := range search(flag.Args()...) {
			fmt.Printf("% 4d. %s\n", i+1, f)
		}
		log.Println("Done.")
	case len(flag.Args()) > 0:
		for _, f := range flag.Args() {
			storeFile(os.Open(f))
		}
		if err := saveFilters(filterFilename); err != nil {
			log.Fatalf("Could not save filters: %s", err)
			os.Exit(1)
		}
	}
}
