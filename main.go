package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/key"

	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer() func() {
	agentURI := os.Getenv("JAEGER_AGENT_URI")
	if agentURI == "" {
		agentURI = "localhost:6831"
	}

	_, flush, err := jaeger.NewExportPipeline(
		jaeger.WithAgentEndpoint(agentURI),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "trace-demo",
			Tags: []core.KeyValue{
				key.String("runtime", "golang"),
				key.String("api", "opentelemetry"),
			},
		}),
		jaeger.RegisterAsGlobal(),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		log.Fatal(err)
	}

	return func() {
		flush()
	}
}

func fibonacci(ctx context.Context, n int) int {
	tr := global.Tracer("fibonacci-handler")
	_, span := tr.Start(ctx, fmt.Sprintf("fibonnaci-%d", n))
	defer span.End()

	switch {
	case n == 1:
		return 0
	case n == 2:
		return 1
	case n > 2:
		return fibonacci(ctx, n-1) + fibonacci(ctx, n-2)
	default:
		return -1
	}
}

func fibonacciHandler(w http.ResponseWriter, r *http.Request) {
	num, _ := strconv.Atoi(mux.Vars(r)["num"])

	ctx := context.Background()
	tracer := global.Tracer("fibonacci-handler")
	ctx, span := tracer.Start(ctx, "handler")
	defer span.End()

	result := fibonacci(ctx, num)

	if result < 0 {
		http.Error(w, "Invalid parameter, needs to be bigger than 0.", http.StatusBadRequest)
		log.Println("Invalid request received!")
		return
	}

	fmt.Fprintf(w, "Fiboncci of %d is %d.\n", num, result)
	log.Println("Valid request received and processed!")
}

func main() {
	initClosure := initTracer()
	defer initClosure()

	router := mux.NewRouter()
	router.HandleFunc("/{num:[0-9]+}", fibonacciHandler)
	log.Println("HTTP server started...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
