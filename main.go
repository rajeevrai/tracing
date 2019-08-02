package main


import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)


func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		url := os.Getenv("DEST_URL")

		resp, err := http.Get(url)

		if err != nil {
			log.Fatalln(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		c.JSON(200, gin.H{
			"message": body,
		})
	})

	r.Run()
}
