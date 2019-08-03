package main

import (
    "bytes"
    "fmt"
    "github.com/gin-gonic/gin"
    opentracing "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "github.com/uber/jaeger-client-go"
    jaegercfg "github.com/uber/jaeger-client-go/config"
    jaegerlog "github.com/uber/jaeger-client-go/log"
    "github.com/uber/jaeger-lib/metrics"
    "io/ioutil"
    "net/url"
    "log"
    "math/rand"
    "net/http"
    "strings"
)

const(
    gimliurlbase = "http://localhost:28090/"
)

func main() {
    r := gin.Default()

    r.GET("/ping", func(c *gin.Context) {
        callGimli(c, "ping")

        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    r.POST("/v1/shorten", func(c *gin.Context) {
        body := string(callGimli(c, "shorten"))

        c.JSON(200, gin.H{
            "message": body,
        })
    })

    r.POST("/v1/shorten_delayed", func(c *gin.Context) {
        body := string(callGimli(c, "shorten_delayed"))

        c.JSON(200, gin.H{
            "message": body,
        })
    })

    r.Run("0.0.0.0:8000")
}

func callGimli(ctx *gin.Context, val string) []byte {
    cfg := jaegercfg.Configuration{
        ServiceName: "Sample Client",
        Sampler: &jaegercfg.SamplerConfig{
            Type:  jaeger.SamplerTypeConst,
            Param: 1,
        },
        Reporter: &jaegercfg.ReporterConfig{
            LogSpans: true,
        },
    }

    // Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
    // and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
    // frameworks.
    jLogger := jaegerlog.StdLogger
    jMetricsFactory := metrics.NullFactory

    // Initialize tracer with a logger and a metrics factory
    tracer, closer, err := cfg.NewTracer(
        jaegercfg.Logger(jLogger),
        jaegercfg.Metrics(jMetricsFactory),
    )
    if err != nil {
        log.Fatalf("Error:%v", err)
    }

    traceID := ctx.Request.Header.Get("X-Trace-ID")
    if traceID == "" {
        traceID = string(rand.Int())
    }

    fmt.Println(fmt.Sprintf("x-trace-id: %s", traceID))

    opentracing.SetGlobalTracer(tracer)
    defer closer.Close()

    clientSpan := tracer.StartSpan("client")
    defer clientSpan.Finish()

    // Set some tags on the clientSpan to annotate that it's the client span. The additional HTTP tags are useful for debugging purposes.
    ext.SpanKindRPCClient.Set(clientSpan)

    var uri string
    var req *http.Request

    jsonStr := []byte(`{"url":"https://www.google.com"}`)

    if val == "ping" {
        ext.HTTPMethod.Set(clientSpan, "GET")

        uri = gimliurlbase + "/ping"
        req, _ = http.NewRequest("GET", uri, nil)
    } else if val == "shorten" {
        ext.HTTPMethod.Set(clientSpan, "POST")

        uri = gimliurlbase + "/v1/shorten"
        req, _ = http.NewRequest("POST", uri, bytes.NewBuffer(jsonStr))
    } else {
        ext.HTTPMethod.Set(clientSpan, "POST")

        uri = gimliurlbase + "/v1/shorten_delayed"
        req, _ = http.NewRequest("POST", uri, bytes.NewBuffer(jsonStr))
    }

    req.Header.Set("X-Trace-ID", traceID)
    ext.HTTPUrl.Set(clientSpan, uri)

    // Inject the client span context into the headers
    tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Fatalf("Client Error:%v\n", err)
    }
    defer resp.Body.Close()

    fmt.Println("response Status:", resp.Status)
    fmt.Println("response Headers:", resp.Header)
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println("response Body:", string(body))

    return body
}
