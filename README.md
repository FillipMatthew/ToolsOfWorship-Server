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
MASTER_KEY=base64_encoded_32_byte_key  # Master key for key encryption in the DB
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
-masterKey    Base64 encoded 32-byte master key for encrypting keys in the DB
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

### Authentication

- POST `/login` - User authentication using basic auth
- POST `/register` - User registration with email verification
- GET `/verifyemail` - Email verification with token
- GET `/health` - Server health check with DB status

## Project Structure

```
├── cmd/
│   └── tow-server/       # Main application entry point and configuration
├── internal/
│   ├── api/             # API handlers and routing
│   │   └── users/       # User-related API handlers
│   ├── config/          # Configuration interfaces
│   ├── db/              # Database implementations
│   │   └── postgresql/  # PostgreSQL specific implementation
│   ├── domain/          # Domain models and interfaces
│   ├── keys/            # Cryptographic operations
│   └── service/         # Business logic services
├── templates/           # HTML/Email templates
└── bin/                # Compiled binaries
```

## Development

### Code Structure

- Uses clean architecture principles
- Separation of concerns between API, domain logic, and data access
- Centralized configuration management
- Secure token handling with key rotation

### Testing

Run tests with (not yet implemented):
```bash
go test ./...
```

### Security Features

- JWT token-based authentication
- Email verification system
- Password hashing using bcrypt
- Key rotation for signing and encryption
- HTTPS/TLS support

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
