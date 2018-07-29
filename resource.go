package govtvfetcher

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

type Resource struct {
	Uri    string
	Length uint64
}

func NewResource(uri string) (*Resource, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, fmt.Errorf("could not parse uri: %v", err)
	}
	clip_id := u.Query().Get("clip_id")
	if clip_id == "" {
		return nil, fmt.Errorf("Could not find clip_id in URI '%s'", uri)
	}

	// Now we have clip_id, we can get the media URI
	infoUri := fmt.Sprintf("http://%s/ASX.php?clip_id=%s", u.Host, clip_id)
	resp, err := http.Get(infoUri)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch uri (%s): %v", infoUri, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("couldn't read body of uri (%s): %v", infoUri, err)
	}
	// response looks like:
	// <ASX version="3.0">
	// <TITLE>City and County of San Francisco</TITLE>
	//   <ENTRY>
	// 	<TITLE>City and County of San Francisco</TITLE>
	// 	<STARTTIME VALUE="00:00:00" />
	// 	<REF HREF="rtmp://207.7.154.100/OnDemand/mp4:sanfrancisco/sanfrancisco_dd1fc377-1168-4e32-a5fe-30be375e5672.mp4?wmcache=0" />
	//   </ENTRY>
	// </ASX>
	// Just find the sanfrancisco_...mp4 portion
	pattern := regexp.MustCompile(`OnDemand/mp4:(([^.]+)\.mp4).*`)
	mediaId := pattern.FindStringSubmatch(string(body))[1]
	if mediaId == "" {
		return nil, fmt.Errorf("couldn't find mediaId: %s", body)
	}
	mediaUri := fmt.Sprintf("http://media-06.granicus.com:443/OnDemand/%s", mediaId)

	resp, err = http.DefaultClient.Head(mediaUri)
	if err != nil {
		return nil, fmt.Errorf("couldn't read uri (%s): %v", mediaUri, err)
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
		Uri:    mediaUri,
		Length: uint64(length),
	}
	return r, nil
}
