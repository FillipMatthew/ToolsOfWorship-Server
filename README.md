# ToolsOfWorship-Server

Tools of Worship server implemented in Go, providing backend services for the Tools of Worship site.

## Features

- User authentication and authorization
- Email verification system
- Secure token management
- PostgreSQL database integration
- HTTPS/TLS support
- Configurable server settings

## Prerequisites

- Go 1.23.1 or higher
- PostgreSQL database
- SMTP server for email functionality

## Installation

1. Clone the repository:
```bash
git clone https://github.com/FillipMatthew/ToolsOfWorship-Server.git
```

2. Navigate to the project directory:
```bash
cd ToolsOfWorship-Server
```

3. Install dependencies:
```bash
go mod download
```

## Configuration

The server can be configured using:
- Environment variables
- Configuration file (config.json)
- Command line flags

### Environment Variables

```env
USE_TLS=true
LISTEN_ADDRESS=:443
CERT_PATH=./certs/cert.pem
KEY_PATH=./certs/key.pem
PUBLIC_DIR=./public
DOMAIN=ToolsOfWorship.com

DB_USE_SSL=true
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=ToW
```

### Command Line Flags

```bash
-tls          Use TLS (default: true)
-address      Listen address (default: :443)
-cert         TLS certificate path
-key          TLS private key path
-public       Public directory path
-domain       Base domain for server endpoints
-dbssl        Use SSL for database
-dbhost       Database host
-dbport       Database port
-dbuser       Database user
-dbpassword   Database password
-dbname       Database name
```

## Building and Running

1. Build the server:
```bash
go build -o bin/tow-server ./cmd/tow-server
```

2. Run the server:
```bash
./bin/tow-server
```

## API Endpoints

- POST `/login` - User authentication
- POST `/register` - User registration
- GET `/verifyemail` - Email verification
- GET `/health` - Server health check

## Project Structure

```
├── cmd/
│   └── tow-server/       # Main application entry point
├── internal/
│   ├── api/             # API handlers and routing
│   ├── config/          # Configuration interfaces
│   ├── db/              # Database implementations
│   ├── domain/          # Domain models and interfaces
│   └── service/         # Business logic services
├── templates/           # Email templates
└── bin/                # Compiled binaries
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
