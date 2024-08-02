package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	err := run(ctx, logger)
	if err != nil {
		logger.ErrorContext(ctx, "unexpected  termination", "err", err.Error())
		os.Exit(1)
	}
}

func run(_ context.Context, log *slog.Logger) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/hi", handleHi())
	mux.HandleFunc("/drop", handleDropped(log, 2*time.Second, false))
	mux.HandleFunc("/dropNoisy", handleDropped(log, 2*time.Second, true))
	mux.HandleFunc("/hang", handleHang(log))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	err := http.ListenAndServe(":"+port, withLogging(log, mux))
	return err
}

func handleHi() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hi!")
	}
}

// handleHang will accept the request and then hang without a response.
func handleHang(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// There is probably a more efficient way to write this and you should probably
		// not take information straight out of request headers and dump
		// it into the logs like this.
		b3TraceID := r.Header.Get("X-B3-TraceId")
		b3SpanID := r.Header.Get("X-B3-SpanId")
		tp := r.Header.Get("traceparent")
		ts := r.Header.Get("tracestate")
		log.Info("waiting indefinitely", "from", r.RemoteAddr, "method", r.Method, "url", r.URL.String(), "traceparent", tp, "tracestate", ts, "X-B3-TraceId", b3TraceID, "X-B3-SpanId", b3SpanID)
		select {}
	}
}

// handleDropped will hijack the connection and then silently drop it after a delay.
// If withResponse is true the handler will respond with a few bytes before dropping
// the connection.
//
// The code is modified from the example given in the Go documentation.
// Source: https://pkg.go.dev/net/http#Hijacker
func handleDropped(log *slog.Logger, after time.Duration, withResponse bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			log.Error("server doesn't support hijacking")
			http.Error(w, "server doesn't support hijacking", http.StatusInternalServerError)
			return
		}

		conn, bufrw, err := hj.Hijack()
		if err != nil {
			log.Error("unable to hijack server", "err", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if withResponse {
			bufrw.WriteString(r.Proto + " 200 OK\n")

			// Note that the content length returned here is deliberately greater
			// than the number of bytes we send in the response body.
			bufrw.WriteString("Content-Length: 3\n\n")
			bufrw.WriteString("Hi")
			bufrw.Flush()
		}

		delay := time.After(after)
		<-delay

		// There is probably a more efficient way to write this and you should probably
		// not take information straight out of request headers and dump
		// it into the logs like this.
		b3TraceID := r.Header.Get("X-B3-TraceId")
		b3SpanID := r.Header.Get("X-B3-SpanId")
		tp := r.Header.Get("traceparent")
		ts := r.Header.Get("tracestate")
		log.Info("dropping connection", "from", r.RemoteAddr, "method", r.Method, "url", r.URL.String(), "traceparent", tp, "tracestate", ts, "X-B3-TraceId", b3TraceID, "X-B3-SpanId", b3SpanID)

		conn.Close()
	}
}

func withLogging(log *slog.Logger, h http.Handler) http.Handler {
	log.Info("using middleware", "function", "withLogging")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// There is probably a more efficient way to write this and you should probably
		// not take information straight out of request headers and dump
		// it into the logs like this.
		b3TraceID := r.Header.Get("X-B3-TraceId")
		b3SpanID := r.Header.Get("X-B3-SpanId")
		tp := r.Header.Get("traceparent")
		ts := r.Header.Get("tracestate")
		log.Info("request", "from", r.RemoteAddr, "method", r.Method, "url", r.URL.String(), "traceparent", tp, "tracestate", ts, "X-B3-TraceId", b3TraceID, "X-B3-SpanId", b3SpanID)

		h.ServeHTTP(w, r)
	})
}
