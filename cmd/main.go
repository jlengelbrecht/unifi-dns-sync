package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"

    "github.com/jlengelbrecht/unifi-dns-sync/internal/handlers"
    "github.com/jlengelbrecht/unifi-dns-sync/internal/store"
)

func main() {
    var (
        port       = flag.Int("port", 52638, "Port to run the server on")
        dataDir    = flag.String("data-dir", "data", "Directory for data storage")
    )
    flag.Parse()

    // Ensure data directory exists
    if err := os.MkdirAll(*dataDir, 0755); err != nil {
        log.Fatalf("Failed to create data directory: %v", err)
    }

    // Initialize database
    dbPath := filepath.Join(*dataDir, "unifi-dns.db")
    store, err := store.NewStore(dbPath)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer store.Close()

    // Initialize handler
    h, err := handlers.NewHandler("web/templates", store)
    if err != nil {
        log.Fatalf("Failed to initialize handler: %v", err)
    }

    // Set up routes with middleware
    http.HandleFunc("/", handlers.Chain(h.Index,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    http.HandleFunc("/setup", handlers.Chain(h.Setup,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    http.HandleFunc("/login", handlers.Chain(h.Login,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    http.HandleFunc("/logout", handlers.Chain(h.Logout,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    http.HandleFunc("/onboarding", handlers.Chain(h.Onboarding,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    http.HandleFunc("/api/devices", handlers.Chain(h.GetDevices,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
        handlers.JSONMiddleware,
        handlers.CORSMiddleware,
    ))
    
    http.HandleFunc("/api/devices/add", handlers.Chain(h.AddDevice,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
        handlers.JSONMiddleware,
        handlers.CORSMiddleware,
    ))

    // Start server
    addr := fmt.Sprintf("0.0.0.0:%d", *port)
    log.Printf("Starting server on %s", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}