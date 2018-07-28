package main

import (
	"fmt"
	"github.com/sbuss/govtvfetcher"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	r, err := govtvfetcher.NewResource("http://media-06.granicus.com:443/OnDemand/sanfrancisco/sanfrancisco_c6d8c565-1eba-41b8-b9c2-f5999a2b141f.mp4")
	if err != nil {
		log.Fatalf("Could not create resource: %v\n", err)
	}

	parts := strings.Split(r.Uri, "/")
	fname := strings.Split(parts[len(parts)-1], ".")[0]
	d, err := ioutil.TempDir("", fname)
	if err != nil {
		log.Fatalf("Could not create tempdir: %v\n", err)
	}
	log.Printf("Saving files to %s\n", d)

	var wg sync.WaitGroup
	chunksize := 1 * 1024 * 1024
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(i int) {
			start := i * chunksize
			stop := (i + 1) * chunksize
			log.Printf("Getting %d offset: %d-%d\n", i, start, stop)
			defer wg.Done()
			bytes, err := r.Get(start, stop)
			if err != nil {
				log.Fatalf("Could not fetch resource: %v\n", err)
			}
			outfile := filepath.Join(d, fmt.Sprintf("%d.mp4", i))
			if err := ioutil.WriteFile(outfile, bytes, 0644); err != nil {
				log.Fatalf("Failed to write bytes: %v", err)
			}
			log.Printf("Wrote file %s\n", outfile)
		}(i)
	}
	wg.Wait()
	fmt.Println("Done")
}
