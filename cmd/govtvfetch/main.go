package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	"github.com/remeh/sizedwaitgroup"
	"github.com/sbuss/govtvfetcher"
)

var (
	keep      = flag.Bool("keep", false, "Keep the temp dir with partial files.")
	chunksize = flag.String("chunksize", "16MB", "Size of video chunks.")
)

func main() {
	flag.Parse()

	r, err := govtvfetcher.NewResource("http://media-06.granicus.com:443/OnDemand/sanfrancisco/sanfrancisco_c6d8c565-1eba-41b8-b9c2-f5999a2b141f.mp4")
	if err != nil {
		log.Fatalf("Could not create resource: %v\n", err)
	}

	parts := strings.Split(r.Uri, "/")
	fname_full := parts[len(parts)-1]
	fname := strings.Split(fname_full, ".")[0]
	d, err := ioutil.TempDir("", fmt.Sprintf("%s-", fname))
	if err != nil {
		log.Fatalf("Could not create tempdir: %v\n", err)
	}
	if !*keep {
		defer os.RemoveAll(d)
	}
	// d := "/tmp/sanfrancisco_c6d8c565-1eba-41b8-b9c2-f5999a2b141f-190550762"
	log.Printf("Saving files to %s\n", d)

	chunksize_bytes, err := bytefmt.ToBytes(*chunksize)
	if err != nil {
		log.Fatalf("Could not convert '%s' to bytes: %v", *chunksize, err)
	}
	num_chunks := (r.Length / chunksize_bytes) + 1
	log.Printf("splitting into %d chunks\n", num_chunks)
	wg := sizedwaitgroup.New(20)
	// Note: the Range header is inclusive on both sides, so two subsequent
	// requests will return overlapping byte ranges, eg:
	// GET(0, 16) + GET(16, 32) returns a total of 33 bytes, with the 16th
	// byte duplicated.
	for i := uint64(0); i < num_chunks; i++ {
		wg.Add()
		go func(i uint64) {
			defer wg.Done()
			start := i * chunksize_bytes
			if start > r.Length {
				log.Printf("start > %d\n", r.Length)
				return
			}
			stop := (i+1)*chunksize_bytes - 1
			if stop > r.Length {
				stop = r.Length
				log.Printf("stop is larger than filesize: %d > %d", stop, r.Length)
			}
			log.Printf("Getting %d offset: %d-%d\n", i, start, stop)
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
	log.Printf("Finished fetching chunks. Combining them into one file: %s", fname_full)

	// Open file for writing
	file, err := os.OpenFile(fname_full, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a buffered writer from the file
	bufferedWriter := bufio.NewWriter(file)

	// Write bytes to buffer

	for i := uint64(0); i < num_chunks; i++ {
		log.Printf("Reading %d\n", i)
		inf := filepath.Join(d, fmt.Sprintf("%d.mp4", i))
		data, err := ioutil.ReadFile(inf)
		if err != nil {
			log.Fatalf("couldn't read file %s: %v\n", inf, err)
		}
		bytesWritten, err := bufferedWriter.Write(data)
		if err != nil {
			log.Fatalf("couldn't write to file %s: %v\n", fname_full, err)
		}
		log.Printf("Wrote %d bytes", bytesWritten)
	}

	fmt.Println("Done")
}
