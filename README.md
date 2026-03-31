# ToolsOfWorship-Server

Tools of Worship server implemented in Go, providing backend services for the Tools of Worship site.

## Features

- User authentication and authorization with JWT
- Email verification system
- Secure token management with key rotation
- PostgreSQL database integration with automatic database creation
- Configurable server settings via environment variables, JSON file, and command-line flags

## Prerequisites

- Go 1.23.1 or higher
- PostgreSQL database
- Mailgun API for email functionality

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
LISTEN_ADDRESS=:8080
DOMAIN=ToolsOfWorship.com

DB_USE_SSL=true
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=ToW
MASTER_KEY=base64_encoded_32_byte_key  # Master key for key encryption in the DB
MAIL_KEY=your_mail_key
```

### Configuration File (config.json)

```json
{
  "server": {
    "address": ":8080",
    "domain": "ToolsOfWorship.com"
  },
  "database": {
    "ssl": true,
    "host": "localhost",
    "port": 5432,
    "user": "user",
    "password": "password",
    "name": "ToW",
    "masterKey": "base64_encoded_32_byte_key"
  },
  "mail": {
    "key": "your_mail_key"
  }
}
```

### NGINX Configuration

To run behind NGINX, use a configuration similar to the following. NGINX will handle SSL termination and serve static files from the `public/` directory.

```nginx
server {
    listen 80;
    server_name toolsofworship.com;

    # Allow static files over HTTP
    location / {
        root /path/to/ToolsOfWorship-Server/public;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # Redirect API calls to HTTPS (optional, but recommended)
    location /api/ {
        return 301 https://$server_name$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name toolsofworship.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # SSL Security Settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA:AES256-GCM-SHA384:AES256-SHA;

    location / {
        root /path/to/ToolsOfWorship-Server/public;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://localhost:8080/health;
    }
}
```

## Deployment

### GitHub Actions

This project uses GitHub Actions for automated deployment. When changes are pushed to the `main` branch, the workflow will:

1. Build the Go binary.
2. Deploy the binary, `public/` directory, and `templates/` directory to the server via SCP.
3. Restart the `tow-server` service via SSH.

To set up deployment, add the following secrets to your GitHub repository:

- `SERVER_HOST`: Server IP or hostname.
- `SERVER_USER`: SSH username.
- `SSH_PRIVATE_KEY`: Private SSH key for deployment.
- `TARGET_DIR`: Directory on the server where the application will be deployed.

### Systemd Service

A sample systemd service file is provided in `tow-server.service`. It is recommended to run the server under a dedicated non-privileged user.

```ini
[Unit]
Description=Tools of Worship Server
After=network.target postgresql.service

[Service]
Type=simple
User=tow
Group=tow
WorkingDirectory=/usr/local/lib/tools-of-worship
ExecStart=/usr/local/lib/tools-of-worship/tow-server
Restart=always
RestartSec=5

# Set environment variables directly
Environment=LISTEN_ADDRESS=:8080
Environment=DOMAIN=toolsofworship.com
Environment=DB_HOST=localhost
Environment=DB_USER=tow_user
Environment=DB_NAME=ToW
# It is better to use an EnvironmentFile for sensitive keys (MASTER_KEY, MAIL_KEY)
# EnvironmentFile=/usr/local/lib/tools-of-worship/.env

# Security Hardening
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=full
ProtectHome=yes

[Install]
WantedBy=multi-user.target
```

To install the service:

1. Create a dedicated user: `sudo useradd -r -s /bin/false tow`
2. Create the working directory and copy the binary:
   ```bash
   sudo mkdir -p /usr/local/lib/tools-of-worship
   sudo cp tow-server /usr/local/lib/tools-of-worship/
   sudo chown -R tow:tow /usr/local/lib/tools-of-worship
   ```
3. Copy the service file: `sudo cp tow-server.service /etc/systemd/system/`
4. Enable and start:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable tow-server
   sudo systemctl start tow-server
   ```

## API Endpoints

### Authentication

- POST `/api/user/login` - User authentication using email and password
- POST `/api/user/register` - User registration with email verification
- GET `/api/user/verifyemail` - Email verification with token
- GET `/health` - Server health check with database status

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

(Note: Ensure database and environment variables are properly configured before running tests.)

### Security Features

- JWT token-based authentication with HMAC-SHA256
- Email verification
- Password hashing using bcrypt for secure storage
- Key rotation for signing and encryption with AES-GCM

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
