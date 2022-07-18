package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-grafana-tempo/pkg/tracer"

	"github.com/gorilla/mux"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	PORT           = 8080
	tempoOtelHost  = "localhost:4317"
	serviceName    = "test-service"
	DefaultAppName = "test-app"
)

type abc struct {
	A string
}

func main() {
	var err error

	shutdownTracer := tracer.NewOtel(tempoOtelHost, serviceName, DefaultAppName)
	defer shutdownTracer()

	r := mux.NewRouter()
	r.Use(otelmux.Middleware(DefaultAppName))

	r.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.StartSpan(r.Context())
		defer span.End()

		time.Sleep(5 * time.Second)

		other(ctx)

		out := abc{A: "x"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(out)
		if err != nil {
			log.Print(err)
			span.SetStatus(codes.Error, "Handler execute error")
			span.RecordError(err)
		}

		span.AddEvent("Users", trace.WithAttributes(attribute.String("ID", out.A)))
		span.SetStatus(codes.Ok, "Handler execute success")
		span.SetAttributes(attribute.String("Test", "Test"))
	}).Methods(http.MethodPost)

	log.Print("Start server in port:", PORT)

	err = http.ListenAndServe(fmt.Sprintf(":%d", PORT), r)
	if err != nil {
		log.Fatalln("Error start server", err)
		return
	}
}

func other(ctx context.Context) {
	_, span := tracer.StartSpan(ctx)
	defer span.End()

	time.Sleep(5 * time.Second)
}
