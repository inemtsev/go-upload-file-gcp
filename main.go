package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"time"
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
	creds, isFound := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !isFound {
		panic("GCP environment variable is not set")
	} else if creds == "" {
		panic("GCP environment variable is empty")
	}

	const bucketName = "el-my-gallery"

	mr, e := c.Request.MultipartReader()
	if e != nil {
		panic("Error reading request:" + e.Error())
	}

	cWithTimeout, cancel := context.WithTimeout(c, time.Second*60)
	defer cancel()
	client, e := storage.NewClient(cWithTimeout)
	if e != nil {
		panic("Error creating client")
	}

	bucket := client.Bucket(bucketName)

	for {
		p, e := mr.NextPart()

		if e == io.EOF {
			break
		} else if e != nil {
			panic("Error processing file:" + e.Error())
		}

		w := bucket.Object(p.FileName()).NewWriter(cWithTimeout)

		if _, e := io.Copy(w, p); e != nil {
			panic("Error during chunk upload:" + e.Error())
		} else if e := w.Close(); e != nil {
			panic("Could not finalize chunk writing:" + e.Error())
		}

		fmt.Printf("Uploaded: %v bytes", w.Size)
	}
}