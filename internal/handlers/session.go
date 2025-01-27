package handlers

import (
    "encoding/base64"
    "net/http"
    "sync"
    "time"

    "github.com/google/uuid"
)

type Session struct {
    ID        string
    UserID    string
    CreatedAt time.Time
    ExpiresAt time.Time
}

type SessionManager struct {
    sessions map[string]*Session
    mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
    return &SessionManager{
        sessions: make(map[string]*Session),
    }
}

func (sm *SessionManager) CreateSession(userID string) *Session {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    session := &Session{
        ID:        base64.URLEncoding.EncodeToString(uuid.New().NodeID()),
        UserID:    userID,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }

    sm.sessions[session.ID] = session
    return session
}

func (sm *SessionManager) GetSession(id string) *Session {
    sm.mu.RLock()
    defer sm.mu.RUnlock()

    session, exists := sm.sessions[id]
    if !exists {
        return nil
    }

    if time.Now().After(session.ExpiresAt) {
        sm.mu.RUnlock()
        sm.DestroySession(id)
        return nil
    }

    return session
}

func (sm *SessionManager) DestroySession(id string) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    delete(sm.sessions, id)
}

func (sm *SessionManager) SetSessionCookie(w http.ResponseWriter, session *Session) {
    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    session.ID,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        Expires:  session.ExpiresAt,
    })
}

func (sm *SessionManager) ClearSessionCookie(w http.ResponseWriter) {
    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        Expires:  time.Now().Add(-24 * time.Hour),
    })
}

func (sm *SessionManager) GetSessionFromRequest(r *http.Request) *Session {
    cookie, err := r.Cookie("session_id")
    if err != nil {
        return nil
    }
    return sm.GetSession(cookie.Value)
}