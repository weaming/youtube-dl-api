package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kkdai/youtube/v2"
	ytdl "github.com/kkdai/youtube/v2/downloader"
	"github.com/olekukonko/tablewriter"
)

var (
	outputFile    string
	outputDir     string
	outputQuality string
	info          bool
)
var priorities = []string{
	// "hd2160",
	// "hd1440",
	"hd1080",
	"hd720",
	"large",
	"medium",
	"small",
	"tiny",
}

func main() {
	usr, _ := user.Current()
	flag.StringVar(&outputDir, "d", filepath.Join(usr.HomeDir, "Downloads", "youtube"), "The output directory.")
	flag.BoolVar(&info, "info", false, "show info of video")
	flag.Parse()

	arg := flag.Arg(0)
	if len(arg) != 0 {
		if dst, e := download(arg); e != nil {
			log.Println(e)
		} else {
			log.Println(dst)
		}
		return
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/download/youtube", func(c *gin.Context) {
		url := c.Query("url")
		if url == "" {
			c.JSON(400, gin.H{
				"error": "missing url",
			})
			return
		}

		destFile, e := download(url)
		if e == nil {
			log.Println(destFile)
			c.File(destFile)
			return
		}
		c.JSON(500, gin.H{
			"error": e.Error(),
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

func download(url string) (string, error) {
	httpTransport := newHTTPTransport()
	dl := ytdl.Downloader{
		OutputDir: outputDir,
	}
	dl.HTTPClient = &http.Client{Transport: httpTransport}

	video, err := dl.GetVideo(url)
	if err != nil {
		return "", err
	}

	if info {
		fmt.Printf("Title:    %s\n", video.Title)
		fmt.Printf("Author:   %s\n", video.Author)
		fmt.Printf("Duration: %v\n", video.Duration)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"itag", "quality", "MimeType"})

		for _, itag := range video.Formats {
			table.Append([]string{strconv.Itoa(itag.ItagNo), itag.Quality, itag.MimeType})
		}
		table.Render()
	}

	var format *youtube.Format
	for _, quality := range priorities {
		if fmt := video.Formats.FindByQuality(quality); fmt != nil && strings.Contains(fmt.MimeType, "mp4") {
			format = fmt
			break
		}
	}
	if format == nil {
		log.Println("format not found!")
		format = &video.Formats[0]
	}
	log.Println("got format:", format.ItagNo, format.Quality, format.MimeType, format.URL)

	log.Println("download to directory", outputDir)
	if format.Quality == "hd1080" {
		fmt.Println("check ffmpeg is installed....")
		ffmpegVersionCmd := exec.Command("ffmpeg", "-version")
		if err := ffmpegVersionCmd.Run(); err != nil {
			return "", fmt.Errorf("please check ffmpeg is installed correctly, err: %w", err)
		}
		return dl.DownloadWithHighQuality(context.Background(), outputFile, video, format.Quality)
	}
	return dl.Download(context.Background(), video, format, outputFile)
}

func newHTTPTransport() *http.Transport {
	httpTransport := &http.Transport{
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpTransport.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	return httpTransport
}
