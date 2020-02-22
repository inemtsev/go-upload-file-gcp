package main

import 	(
	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"google.golang.org/api/option"
	"io"
	"log"
)

 const uploadBucket = "eventslooped-media"
 const uploadApiKey = "AIzaSyB8fOy8GGW8n9Hdquu5bLfHbKAY2fWeRA8"

func main() {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:           []string{"http://localhost:3000"},
		AllowMethods:           []string{"PATCH"},
	}))

	rg := router.Group("api/v1/photo")
	{
		rg.PATCH("/", uploadFile)
	}

	router.Run()
}

func uploadFile(c *gin.Context) {
	mr, e := c.Request.MultipartReader()
	if e != nil {
		panic("Error reading request")
	}

	client, e := storage.NewClient(c, option.WithAPIKey(uploadApiKey))
	bucket := client.Bucket(uploadBucket)

	log.Println("Client initialized...")

	for {
		p, e := mr.NextPart()
		log.Println("Reading part...")

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