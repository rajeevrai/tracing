package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	utils "github.com/razorpay/goutils"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		url := os.Getenv("DEST_URL")

		requestID := generateRequestId()
		externalRequestID := c.GetHeader(utils.ExternalRequestID)

		g := gin.Context{}
		g.Set(utils.RequestID, requestID)
		g.Set(utils.ExternalRequestID, externalRequestID)

		ctx := context.Background()
		ctx = context.WithValue(ctx, "gin", &g)

		utils.Init()
		lgr := utils.Get(ctx)

		fields := map[string]interface{}{
			"xxx": "fff",
		}

		// make http request
		client := &http.Client{}
		request, err := http.NewRequest("GET", url, nil)

		if err != nil {
			lgr.Fatal(fmt.Sprintf("no request: %s", err.Error()), fields)
		}

		// add headers
		request.Header.Set(utils.ExternalRequestID, externalRequestID)
		request.Header.Set(utils.RequestID, requestID)

		resp, err := client.Do(request)

		if err != nil {
			lgr.Fatal(fmt.Sprintf("error from http request: %s", err.Error()), fields)
		}

		lgr.Trace("GIMLI_CALL", fields)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			lgr.Fatal(fmt.Sprintf("error reading request body request: %s", err.Error()), fields)
		}

		c.JSON(200, gin.H{
			"message": body,
		})
	})

	r.Run()
}

func generateRequestId() string {
	return "mera-id"
}
