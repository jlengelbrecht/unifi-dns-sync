package handlers

import (
    "log"
    "net/http"
    "runtime/debug"
    "time"
)

type statusWriter struct {
    http.ResponseWriter
    status int
    length int
}

func (w *statusWriter) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
    if w.status == 0 {
        w.status = 200
    }
    n, err := w.ResponseWriter.Write(b)
    w.length += n
    return n, err
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        sw := &statusWriter{ResponseWriter: w}

        next.ServeHTTP(sw, r)

        log.Printf(
            "%s %s %d %s %d bytes %v",
            r.RemoteAddr,
            r.Method,
            sw.status,
            r.URL.Path,
            sw.length,
            time.Since(start),
        )
    }
}

func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v\n%s", err, debug.Stack())
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    }
}

func JSONMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" && r.URL.Path != "/favicon.ico" {
            w.Header().Set("Content-Type", "application/json")
        }
        next.ServeHTTP(w, r)
    }
}

func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    }
}

func Chain(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
    for _, m := range middlewares {
        h = m(h)
    }
    return h
}
