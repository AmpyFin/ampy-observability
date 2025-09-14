package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"        // <-- add this
	"os/signal"
	"time"

	"ampy.local/ampy-observability/sdk/go/ampyobs"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	obs, err := ampyobs.Init(ctx, ampyobs.Config{
		ServiceName:    "ampy-demo-svc",
		ServiceVersion: "0.1.0",
		Environment:    "dev",
		CollectorGRPC:  "127.0.0.1:4317",
	})
	if err != nil {
		log.Fatalf("init obs: %v", err)
	}
	defer obs.Shutdown(context.Background())

	reqs := obs.Metrics.NewCounter("ampy_demo", "requests_total", "Total demo requests.", nil)
	lat := obs.Metrics.NewHistogram("ampy_demo", "request_latency_ms", "Latency in ms.", []float64{1, 2, 5, 10, 20, 50, 100, 200, 500}, nil)

	mux := http.NewServeMux()
	// Register metrics on the same mux the server uses:
	mux.Handle("/metrics", obs.Metrics.Handler())
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = ampyobs.WithDomainContext(ctx, ampyobs.DomainContext{
			RunID:   fmt.Sprintf("demo_run_%d", time.Now().Unix()),
			AsOfISO: time.Now().UTC().Format(time.RFC3339),
			Symbol:  "AAPL",
			MIC:     "XNAS",
		})

		tr := obs.Tracer("ampy-demo")
		_, span := tr.Start(ctx, "work.do")
		defer span.End()

		start := time.Now()
		time.Sleep(time.Duration(5+rand.Intn(60)) * time.Millisecond)

		reqs.With(prometheus.Labels{"domain": "signals", "outcome": "ok", "reason": "none"}).Inc()
		lat.With(prometheus.Labels{"domain": "signals"}).Observe(float64(time.Since(start).Milliseconds()))

		obs.Logger.Info(ctx, "did some work")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	srv := &http.Server{
		Addr:    ":9464",
		Handler: ampyobs.HTTPServerMiddleware(obs)(mux),
	}

	go func() {
		log.Println("demo: serving on http://localhost:9464  (GET /work, /metrics)")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	_ = srv.Shutdown(context.Background())
	fmt.Println("bye")
}
