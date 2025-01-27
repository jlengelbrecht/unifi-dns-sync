package store

import (
    "database/sql"
    "errors"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "golang.org/x/crypto/bcrypt"

    "unifi-dns-sync/internal/models"
)

var (
    ErrNotFound = errors.New("not found")
    ErrExists   = errors.New("already exists")
)

type Store struct {
    db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    store := &Store{db: db}
    if err := store.initSchema(); err != nil {
        return nil, err
    }

    return store, nil
}

func (s *Store) initSchema() error {
    schema := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        username TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        is_admin BOOLEAN NOT NULL,
        created_at DATETIME NOT NULL
    );

    CREATE TABLE IF NOT EXISTS unifi_credentials (
        id TEXT PRIMARY KEY,
        username TEXT NOT NULL,
        password TEXT NOT NULL,
        is_global BOOLEAN NOT NULL,
        created_at DATETIME NOT NULL,
        created_by TEXT NOT NULL,
        FOREIGN KEY(created_by) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS unifi_devices (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        address TEXT NOT NULL,
        created_at DATETIME NOT NULL,
        created_by TEXT NOT NULL,
        use_global BOOLEAN NOT NULL,
        credentials_id TEXT,
        FOREIGN KEY(created_by) REFERENCES users(id),
        FOREIGN KEY(credentials_id) REFERENCES unifi_credentials(id)
    );

    CREATE TABLE IF NOT EXISTS dns_records (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        rrtype TEXT NOT NULL,
        value TEXT NOT NULL,
        device_id TEXT NOT NULL,
        enabled BOOLEAN NOT NULL,
        description TEXT,
        created_at DATETIME NOT NULL,
        updated_at DATETIME NOT NULL,
        created_by TEXT NOT NULL,
        FOREIGN KEY(device_id) REFERENCES unifi_devices(id),
        FOREIGN KEY(created_by) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS app_config (
        is_initialized BOOLEAN NOT NULL,
        global_creds_id TEXT,
        FOREIGN KEY(global_creds_id) REFERENCES unifi_credentials(id)
    );`

    _, err := s.db.Exec(schema)
    return err
}

func (s *Store) CreateUser(user *models.User, password string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    user.PasswordHash = string(hash)
    user.CreatedAt = time.Now()

    _, err = s.db.Exec(
        "INSERT INTO users (id, username, password_hash, is_admin, created_at) VALUES (?, ?, ?, ?, ?)",
        user.ID, user.Username, user.PasswordHash, user.IsAdmin, user.CreatedAt,
    )
    return err
}

func (s *Store) GetUser(username string) (*models.User, error) {
    var user models.User
    err := s.db.QueryRow(
        "SELECT id, username, password_hash, is_admin, created_at FROM users WHERE username = ?",
        username,
    ).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsAdmin, &user.CreatedAt)

    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    return &user, err
}

func (s *Store) ValidateUser(username, password string) (*models.User, error) {
    user, err := s.GetUser(username)
    if err != nil {
        return nil, err
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return nil, errors.New("invalid password")
    }

    return user, nil
}

func (s *Store) CreateDevice(device *models.UnifiDevice) error {
    device.CreatedAt = time.Now()

    _, err := s.db.Exec(
        "INSERT INTO unifi_devices (id, name, address, created_at, created_by, use_global, credentials_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
        device.ID, device.Name, device.Address, device.CreatedAt, device.CreatedBy, device.UseGlobal,
        device.Credentials.ID,
    )
    return err
}

func (s *Store) GetDevice(id string) (*models.UnifiDevice, error) {
    var device models.UnifiDevice
    var credsID sql.NullString

    err := s.db.QueryRow(
        "SELECT id, name, address, created_at, created_by, use_global, credentials_id FROM unifi_devices WHERE id = ?",
        id,
    ).Scan(&device.ID, &device.Name, &device.Address, &device.CreatedAt, &device.CreatedBy,
        &device.UseGlobal, &credsID)

    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, err
    }

    if credsID.Valid {
        creds, err := s.GetCredentials(credsID.String)
        if err != nil {
            return nil, err
        }
        device.Credentials = creds
    }

    return &device, nil
}

func (s *Store) ListDevices() ([]*models.UnifiDevice, error) {
    rows, err := s.db.Query(
        "SELECT id, name, address, created_at, created_by, use_global, credentials_id FROM unifi_devices",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var devices []*models.UnifiDevice
    for rows.Next() {
        var device models.UnifiDevice
        var credsID sql.NullString
        
        if err := rows.Scan(&device.ID, &device.Name, &device.Address, &device.CreatedAt,
            &device.CreatedBy, &device.UseGlobal, &credsID); err != nil {
            return nil, err
        }

        if credsID.Valid {
            creds, err := s.GetCredentials(credsID.String)
            if err != nil {
                return nil, err
            }
            device.Credentials = creds
        }

        devices = append(devices, &device)
    }

    return devices, nil
}

func (s *Store) CreateCredentials(creds *models.UnifiCredentials) error {
    creds.CreatedAt = time.Now()

    _, err := s.db.Exec(
        "INSERT INTO unifi_credentials (id, username, password, is_global, created_at, created_by) VALUES (?, ?, ?, ?, ?, ?)",
        creds.ID, creds.Username, creds.Password, creds.IsGlobal, creds.CreatedAt, creds.CreatedBy,
    )
    return err
}

func (s *Store) GetCredentials(id string) (*models.UnifiCredentials, error) {
    var creds models.UnifiCredentials
    err := s.db.QueryRow(
        "SELECT id, username, password, is_global, created_at, created_by FROM unifi_credentials WHERE id = ?",
        id,
    ).Scan(&creds.ID, &creds.Username, &creds.Password, &creds.IsGlobal, &creds.CreatedAt, &creds.CreatedBy)

    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    return &creds, err
}

func (s *Store) GetGlobalCredentials() (*models.UnifiCredentials, error) {
    var creds models.UnifiCredentials
    err := s.db.QueryRow(
        "SELECT id, username, password, is_global, created_at, created_by FROM unifi_credentials WHERE is_global = true",
    ).Scan(&creds.ID, &creds.Username, &creds.Password, &creds.IsGlobal, &creds.CreatedAt, &creds.CreatedBy)

    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    return &creds, err
}

func (s *Store) GetAppConfig() (*models.AppConfig, error) {
    var config models.AppConfig
    var globalCredsID sql.NullString

    err := s.db.QueryRow(
        "SELECT is_initialized, global_creds_id FROM app_config LIMIT 1",
    ).Scan(&config.IsInitialized, &globalCredsID)

    if err == sql.ErrNoRows {
        // Initialize with defaults
        config.IsInitialized = false
        if err := s.SaveAppConfig(&config); err != nil {
            return nil, err
        }
        return &config, nil
    }
    if err != nil {
        return nil, err
    }

    if globalCredsID.Valid {
        config.GlobalCreds, err = s.GetCredentials(globalCredsID.String)
        if err != nil {
            return nil, err
        }
    }

    return &config, nil
}

func (s *Store) SaveAppConfig(config *models.AppConfig) error {
    var globalCredsID sql.NullString
    if config.GlobalCreds != nil {
        globalCredsID.String = config.GlobalCreds.ID
        globalCredsID.Valid = true
    }

    _, err := s.db.Exec(
        "INSERT OR REPLACE INTO app_config (is_initialized, global_creds_id) VALUES (?, ?)",
        config.IsInitialized, globalCredsID,
    )
    return err
}

func (s *Store) Close() error {
    return s.db.Close()
}