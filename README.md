# Install

go get github.com/sbuss/govtvfetcher/cmd/govtvfetch


# Use

Find the video you want from sfgovtv.org, eg "http://sanfrancisco.granicus.com/MediaPlayer.php?view_id=20&clip_id=31014".
Then just call the fetched:

```sh
govtvfetch -uri 'http://sanfrancisco.granicus.com/MediaPlayer.php?view_id=20&clip_id=31014'
```

If will download the file in chunks to /tmp and then combine them into your
current directory.
