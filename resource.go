package govtvfetcher

import (
	"fmt"
	"net/http"
	"strconv"
)

type Resource struct {
	Uri    string
	Length uint64
}

func NewResource(uri string) (*Resource, error) {
	resp, err := http.DefaultClient.Head(uri)
	if err != nil {
		return nil, fmt.Errorf("couldn't read uri (%s): %v", uri, err)
	}
	want := "video/mp4"
	if got := resp.Header.Get("Content-Type"); got != want {
		return nil, fmt.Errorf("invalid Content-Type header, want: %s, got: %s", got, want)
	}
	length_s := resp.Header.Get("Content-Length")
	if length_s == "" {
		return nil, fmt.Errorf("invalid Content-Length: %s", length_s)
	}
	length, err := strconv.Atoi(length_s)
	if err != nil {
		return nil, fmt.Errorf("invalid length (%s): %v", length_s, err)
	}
	if length < 0 {
		return nil, fmt.Errorf("length must be >= 0, got %d", length)
	}
	r := &Resource{
		Uri:    uri,
		Length: uint64(length),
	}
	return r, nil
}
