package govtvfetcher

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Resource struct {
	Uri    string
	Length uint64
}

func asx(uri string) (string, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return "", fmt.Errorf("could not parse uri: %v", err)
	}
	clip_id := u.Query().Get("clip_id")
	if clip_id == "" {
		return "", fmt.Errorf("Could not find clip_id in URI '%s'", uri)
	}
	view_id := u.Query().Get("view_id")
	if view_id == "" {
		return "", fmt.Errorf("Could not find view_id in URI '%s'", uri)
	}

	// Now we have clip_id, we can get the media URI
	infoUri := fmt.Sprintf("http://%s/ASX.php?view_id=%s&clip_id=%s", u.Host, view_id, clip_id)
	resp, err := http.Get(infoUri)
	if err != nil {
		return "", fmt.Errorf("couldn't fetch uri (%s): %v", infoUri, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't read body of uri (%s): %v", infoUri, err)
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
		return "", fmt.Errorf("couldn't find mediaId: %s", body)
	}
	mediaUri := fmt.Sprintf("http://archive-media.granicus.com:443/OnDemand/%s", mediaId)
	return mediaUri, nil
}

// NewResource takes any of the common media urls and returns a Resource
//
// The common urls look like:
//   * http://sanfrancisco.granicus.com/DownloadFile.php?view_id=10&clip_id=30229
//     via the RSS feed http://sanfrancisco.granicus.com/ViewPublisherRSS.php?view_id=10
//   * http://sanfrancisco.granicus.com/MediaPlayer.php?view_id=10&clip_id=31040
//     via the Board/Commission page http://sanfrancisco.granicus.com/ViewPublisher.php?%20%20%20%20view_id=10
//   * media-06.granicus.com:443/OnDemand/sanfrancisco/sanfrancisco_c6d8c565-1eba-41b8-b9c2-f5999a2b141f.mp4
//     via some manual inspection of resources that the playback page fetches
//   * http://archive-media.granicus.com:443/OnDemand/sanfrancisco/sanfrancisco_618361c0-0a52-4ab0-ae84-e707fc02bc2b.mp4
//     via more manual inspection
// All of these eventually get us to a url which looks like
// http://archive-media.granicus.com:443/OnDemand/sanfrancisco/sanfrancisco_c6d8c565-1eba-41b8-b9c2-f5999a2b141f.mp4
func NewResource(uri string) (*Resource, error) {
	var mediaUri string
	var err error
	if strings.Contains(uri, "DownloadFile") || strings.Contains(uri, "MediaPlayer") {
		mediaUri, err = asx(uri)
		if err != nil {
			return nil, fmt.Errorf("asx: %v", err)
		}
	} else if strings.Contains(uri, ".mp4") {
		mediaUri = uri
	}

	resp, err := http.DefaultClient.Head(mediaUri)
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
