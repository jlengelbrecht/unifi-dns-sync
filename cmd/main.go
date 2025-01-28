package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"

    "github.com/jlengelbrecht/unifi-dns-sync/internal/handlers"
    "github.com/jlengelbrecht/unifi-dns-sync/internal/store"
)

var (
    Version = "dev"
    Commit  = "unknown"
)

func main() {
    var (
        port       = flag.Int("port", 52638, "Port to run the server on")
        dataDir    = flag.String("data-dir", "data", "Directory for data storage")
        debug      = flag.Bool("debug", false, "Enable debug logging")
        templatesDir = flag.String("templates-dir", "../web/templates", "Directory containing HTML templates")
    )
    flag.Parse()

    // Configure logging
    if *debug {
        log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
    } else {
        log.SetFlags(log.Ldate | log.Ltime)
    }

    log.Printf("Starting Unifi DNS Manager %s (%s)", Version, Commit)

    // Create mux for better route handling
    mux := http.NewServeMux()

    // Health check endpoint (no middleware)
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "ok",
            "version": Version,
            "commit":  Commit,
        })
    })

    // Ensure data directory exists with correct permissions
    dataPath := filepath.Join(*dataDir)
    if err := os.MkdirAll(dataPath, 0777); err != nil {
        log.Fatalf("Failed to create data directory: %v", err)
    }

    // Initialize database
    dbPath := filepath.Join(dataPath, "unifi-dns.db")
    store, err := store.NewStore(dbPath)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer store.Close()

    // Initialize handler
    h, err := handlers.NewHandler(*templatesDir, store)
    if err != nil {
        log.Fatalf("Failed to initialize handler: %v", err)
    }

    // Set up routes with middleware
    mux.HandleFunc("/", handlers.Chain(h.Index,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    mux.HandleFunc("/setup", handlers.Chain(h.Setup,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    mux.HandleFunc("/login", handlers.Chain(h.Login,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    mux.HandleFunc("/logout", handlers.Chain(h.Logout,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    mux.HandleFunc("/onboarding", handlers.Chain(h.Onboarding,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
    ))
    
    mux.HandleFunc("/api/devices", handlers.Chain(h.GetDevices,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
        handlers.JSONMiddleware,
        handlers.CORSMiddleware,
    ))
    
    mux.HandleFunc("/api/devices/add", handlers.Chain(h.AddDevice,
        handlers.LoggingMiddleware,
        handlers.RecoveryMiddleware,
        handlers.JSONMiddleware,
        handlers.CORSMiddleware,
    ))

    // Start server
    addr := fmt.Sprintf("0.0.0.0:%d", *port)
    log.Printf("Starting server on %s", addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
