package main

import (
	"cloud.google.com/go/storage"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With"}
	router.Use(cors.New(config))

	rg := router.Group("api/v1")
	{
		rg.POST("/photo", uploadFile)
	}

	router.Run()
}

func uploadFile(c *gin.Context) {
	var f *os.File
	file, header, e := c.Request.FormFile("file")

	if f == nil {
		f, e = os.OpenFile(header.Filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModeAppend)
		if e != nil {
			panic("Error creating file on the filesystem: " + e.Error())
		}
		defer f.Close()
	}

	if _, e := io.Copy(f, file); e != nil {
		panic("Error during chunk write:" + e.Error())
	}

	if isFileUploadCompleted(c) {
		uploadToGoogle(c, f)
	}
}

func isFileUploadCompleted(c *gin.Context) bool {
	contentRangeHeader := c.Request.Header.Get("Content-Range")
	rangeAndSize := strings.Split(contentRangeHeader, "/")
	rangeParts := strings.Split(rangeAndSize[0], "-")

	rangeMax, e := strconv.Atoi(rangeParts[1])
	if e != nil {
		panic("Could not parse range max from header")
	}

	fileSize, e := strconv.Atoi(rangeAndSize[1])
	if e != nil {
		panic("Could not parse file size from header")
	}

	return fileSize == rangeMax
}

func uploadToGoogle(c *gin.Context, f *os.File) {
	creds, isFound := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !isFound {
		panic("GCP environment variable is not set")
	} else if creds == "" {
		panic("GCP environment variable is empty")
	}
	const bucketName = "el-my-gallery"

	client, e := storage.NewClient(c)
	if e != nil {
		panic("Error creating client: " + e.Error())
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	w := bucket.Object(f.Name()).NewWriter(c)
	defer w.Close()

	fmt.Println()
	fmt.Printf("%v is the upload filename", f.Name())
	fmt.Println()

	f.Seek(0, io.SeekStart)
	if bw, e := io.Copy(w, f); e != nil {
		panic("Error during GCP upload:" + e.Error())
	} else {
		fmt.Printf("%v bytes written to Cloud", bw)
		fmt.Println()
	}
}