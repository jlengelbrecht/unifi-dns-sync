package models

import (
    "time"
)

type User struct {
    ID           string    `json:"id"`
    Username     string    `json:"username"`
    PasswordHash string    `json:"-"`
    IsAdmin      bool      `json:"is_admin"`
    CreatedAt    time.Time `json:"created_at"`
}

type UnifiDevice struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Address     string    `json:"address"`
    CreatedAt   time.Time `json:"created_at"`
    CreatedBy   string    `json:"created_by"`
    UseGlobal   bool      `json:"use_global"`
    Credentials *UnifiCredentials `json:"credentials,omitempty"`
}

type UnifiCredentials struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Password  string    `json:"password"`
    IsGlobal  bool      `json:"is_global"`
    CreatedAt time.Time `json:"created_at"`
    CreatedBy string    `json:"created_by"`
}

type DNSRecord struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    RRType      string    `json:"rrtype"`
    Value       string    `json:"value"`
    DeviceID    string    `json:"device_id"`
    Enabled     bool      `json:"enabled"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    CreatedBy   string    `json:"created_by"`
}

type AppConfig struct {
    IsInitialized bool              `json:"is_initialized"`
    GlobalCreds   *UnifiCredentials `json:"global_creds,omitempty"`
}