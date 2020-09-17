# Youtube DL API

```
go get -u github.com/weaming/youtube-dl-api

# download directly
youtube-dl-api -info https://www.youtube.com/watch?v=0Lna01wg-3c

# or run as HTTP service on localhost:8080
youtube-dl-api -info
curl 'localhost:8080/download/youtube?url=https://www.youtube.com/watch?v=TYvRLBiN6Vg' -o name.mp4
```

Highest quality is `hd720`: `hd1080` needs `ffmpeg` to merge vidoe and audio, which consumes too much memory.
