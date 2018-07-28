// Package govtvfetcher provides tools for getting granicus videos efficiently.
package govtvfetcher

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func (r *Resource) Get(start, stop int) ([]byte, error) {
	log.Printf("Attempting to fetch offset %d-%d\n", start, stop)
	if start > stop {
		return nil, fmt.Errorf("start (%d) > stop (%d)", start, stop)
	}
	if stop > r.Length {
		log.Printf("stop is larger than filesize: %d > %d", stop, r.Length)
		stop = r.Length
	}
	if start < 0 {
		return nil, fmt.Errorf("start must be >= 0")
	}
	req, err := http.NewRequest("GET", r.Uri, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make new request: %v", err)
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, stop))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not do request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response: %v", err)
	}
	log.Printf("Successfully fetched offset %d-%d\n", start, stop)
	return body, nil
}
