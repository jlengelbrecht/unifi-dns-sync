package handlers

import (
    "encoding/json"
    "html/template"
    "net/http"
    "path/filepath"

    "github.com/google/uuid"

    "unifi-dns-manager/internal/api"
    "unifi-dns-manager/internal/models"
    "unifi-dns-manager/internal/store"
)

type Handler struct {
    templates      *template.Template
    store         *store.Store
    sessionManager *SessionManager
    clients       map[string]*api.UnifiClient
}

func NewHandler(templatesDir string, store *store.Store) (*Handler, error) {
    tmpl, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
    if err != nil {
        return nil, err
    }

    return &Handler{
        templates:      tmpl,
        store:         store,
        sessionManager: NewSessionManager(),
        clients:       make(map[string]*api.UnifiClient),
    }, nil
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session := h.sessionManager.GetSessionFromRequest(r)
        if session == nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        next(w, r)
    }
}

func (h *Handler) requireInit(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        config, err := h.store.GetAppConfig()
        if err != nil {
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }

        if !config.IsInitialized {
            http.Redirect(w, r, "/setup", http.StatusSeeOther)
            return
        }
        next(w, r)
    }
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }

    config, err := h.store.GetAppConfig()
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    if !config.IsInitialized {
        http.Redirect(w, r, "/setup", http.StatusSeeOther)
        return
    }

    session := h.sessionManager.GetSessionFromRequest(r)
    if session == nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    devices, err := h.store.ListDevices()
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    data := struct {
        Devices []*models.UnifiDevice
    }{
        Devices: devices,
    }

    h.templates.ExecuteTemplate(w, "index.html", data)
}

func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
    config, err := h.store.GetAppConfig()
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    if config.IsInitialized {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    if r.Method == "GET" {
        h.templates.ExecuteTemplate(w, "setup.html", nil)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }

    // Create admin user
    adminUser := &models.User{
        ID:       uuid.New().String(),
        Username: r.FormValue("username"),
        IsAdmin:  true,
    }

    if err := h.store.CreateUser(adminUser, r.FormValue("password")); err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

    // Mark app as initialized
    config.IsInitialized = true
    if err := h.store.SaveAppConfig(config); err != nil {
        http.Error(w, "Failed to save config", http.StatusInternalServerError)
        return
    }

    // Create session and redirect to device setup
    session := h.sessionManager.CreateSession(adminUser.ID)
    h.sessionManager.SetSessionCookie(w, session)
    http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        h.templates.ExecuteTemplate(w, "login.html", nil)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }

    user, err := h.store.ValidateUser(r.FormValue("username"), r.FormValue("password"))
    if err != nil {
        h.templates.ExecuteTemplate(w, "login.html", map[string]string{
            "Error": "Invalid username or password",
        })
        return
    }

    session := h.sessionManager.CreateSession(user.ID)
    h.sessionManager.SetSessionCookie(w, session)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    session := h.sessionManager.GetSessionFromRequest(r)
    if session != nil {
        h.sessionManager.DestroySession(session.ID)
    }
    h.sessionManager.ClearSessionCookie(w)
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) Onboarding(w http.ResponseWriter, r *http.Request) {
    session := h.sessionManager.GetSessionFromRequest(r)
    if session == nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    if r.Method == "GET" {
        h.templates.ExecuteTemplate(w, "onboarding.html", nil)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }

    useGlobalCreds := r.FormValue("use_global") == "true"
    
    // Create credentials
    creds := &models.UnifiCredentials{
        ID:        uuid.New().String(),
        Username:  r.FormValue("username"),
        Password:  r.FormValue("password"),
        IsGlobal:  useGlobalCreds,
        CreatedBy: session.UserID,
    }

    if err := h.store.CreateCredentials(creds); err != nil {
        http.Error(w, "Failed to save credentials", http.StatusInternalServerError)
        return
    }

    // If using global credentials, update app config
    if useGlobalCreds {
        config, err := h.store.GetAppConfig()
        if err != nil {
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }

        config.GlobalCreds = creds
        if err := h.store.SaveAppConfig(config); err != nil {
            http.Error(w, "Failed to save config", http.StatusInternalServerError)
            return
        }
    }

    // Create device
    device := &models.UnifiDevice{
        ID:          uuid.New().String(),
        Name:        r.FormValue("device_name"),
        Address:     r.FormValue("device_address"),
        UseGlobal:   useGlobalCreds,
        CreatedBy:   session.UserID,
        Credentials: creds,
    }

    if err := h.store.CreateDevice(device); err != nil {
        http.Error(w, "Failed to save device", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

// API endpoints for managing devices and DNS records
func (h *Handler) GetDevices(w http.ResponseWriter, r *http.Request) {
    devices, err := h.store.ListDevices()
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(devices)
}

func (h *Handler) AddDevice(w http.ResponseWriter, r *http.Request) {
    session := h.sessionManager.GetSessionFromRequest(r)
    if session == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var device models.UnifiDevice
    if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    device.ID = uuid.New().String()
    device.CreatedBy = session.UserID

    if err := h.store.CreateDevice(&device); err != nil {
        http.Error(w, "Failed to save device", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(device)
}