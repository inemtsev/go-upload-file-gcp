package main

import 	(
	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"io"
)

 const uploadBucket = "eventslooped-media"
 const uploadApiKey = "AIzaSyB8fOy8GGW8n9Hdquu5bLfHbKAY2fWeRA8"

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	router := gin.Default()

	rg := router.Group("api/v1/photo")
	{
		rg.POST("/", uploadFile)
	}

	router.Use(CORSMiddleware())

	router.Run()
}

func uploadFile(c *gin.Context) {
	mr, e := c.Request.MultipartReader()
	if e != nil {
		panic("Error reading request")
	}

	client, e := storage.NewClient(c, option.WithAPIKey(uploadApiKey))
	bucket := client.Bucket(uploadBucket)

	for {
		p, e := mr.NextPart()

		if e == io.EOF {
			break
		} else if e != nil {
			panic("Error processing file")
		}

		w := bucket.Object(p.FileName()).NewWriter(c)

		if _, e := io.Copy(w, p); e != nil {
			panic("Error during chunk upload")
		} else if e := w.Close(); e != nil {
			panic("Could not finalize chunk writing")
		}

	}
}