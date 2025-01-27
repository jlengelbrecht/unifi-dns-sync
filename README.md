# Unifi DNS Manager

A web-based application for managing DNS records across multiple Unifi devices (UDM, UDM Pro, UDM SE, etc.).

## Features

- Web-based interface for managing DNS records
- Support for multiple Unifi devices
- User authentication and management
- Global or per-device credentials
- Docker support for easy deployment
- Persistent storage of settings and credentials
- Modern, responsive UI

## Quick Start with Docker

1. Clone the repository:
```bash
git clone https://github.com/yourusername/unifi-dns-manager.git
cd unifi-dns-manager
```

2. Start the application using Docker Compose:
```bash
docker-compose up -d
```

3. Access the web interface at `http://localhost:52638`

4. Follow the setup wizard to:
   - Create your admin account
   - Add your first Unifi device
   - Configure device credentials

## Manual Installation

### Prerequisites

- Go 1.21 or later
- SQLite3
- Git

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/yourusername/unifi-dns-manager.git
cd unifi-dns-manager
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o unifi-dns-manager ./cmd/main.go
```

4. Run the application:
```bash
./unifi-dns-manager
```

## Configuration

The application uses SQLite for storage, which is created automatically in the `data` directory. When running in Docker, this directory is persisted through a volume mount.

### First-Time Setup

1. Access the web interface
2. Create your admin account
3. Add your first Unifi device:
   - Provide device name and address
   - Configure credentials (global or device-specific)
   - Test the connection

### Adding Additional Devices

1. Log in to the web interface
2. Click "Add Device" in the navigation
3. Fill in the device details:
   - Name and address
   - Choose to use global credentials or device-specific ones
   - Test the connection

## Security Considerations

- All passwords are hashed before storage
- HTTPS is recommended for production use
- Session cookies are secure and HTTP-only
- Credentials are stored in an encrypted SQLite database

## Development

### Project Structure

```
.
├── cmd/
│   └── main.go           # Application entry point
├── internal/
│   ├── api/             # Unifi API client
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   └── store/           # Database operations
├── web/
│   ├── static/          # Static assets
│   └── templates/       # HTML templates
├── Dockerfile          # Docker build instructions
├── docker-compose.yml  # Docker Compose configuration
└── README.md          # This file
```

### Running Tests

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details