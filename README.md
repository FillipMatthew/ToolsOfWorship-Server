# ToolsOfWorship-Server

Go REST API server providing backend services for the Tools of Worship application.

## Features

- JWT authentication with automatic signing key rotation
- Email verification flow via Mailgun
- Password hashing with bcrypt; encryption keys stored AES-GCM encrypted at rest
- PostgreSQL with automatic database creation and SQL migration system
- In-process TTL caching for read-heavy store operations
- Rate limiting on authentication endpoints
- Per-request body size limits per endpoint
- Security headers (HSTS, CSP, X-Frame-Options, etc.)
- CORS with configurable origin allowlist
- Request timeouts on all endpoints
- Structured logging via `log/slog`
- Prometheus metrics at `/metrics` (restricted — see nginx configuration)
- Configurable via environment variables, JSON file, and CLI flags
- OpenAPI 3.0 specification at `api/openapi.yaml`
- Automated deployment via GitHub Actions

## Prerequisites

- Go 1.23.1 or higher
- PostgreSQL
- Mailgun account

## Installation

```bash
git clone https://github.com/FillipMatthew/ToolsOfWorship-Server.git
cd ToolsOfWorship-Server
go mod download
```

## Configuration

Settings are applied in this priority order (highest wins): CLI flags → config.json → environment variables → defaults.

### Environment Variables

```env
LISTEN_ADDRESS=:8080
DOMAIN=ToolsOfWorship.com
VERIFICATION_EMAIL_TEMPLATE_PATH=./templates/VerificationEmailTemplate.html
CORS_ALLOWED_ORIGINS=https://example.com,https://www.example.com  # empty = wildcard
REQUEST_TIMEOUT_SECS=30

DB_USE_SSL=true
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=ToW
MASTER_KEY=base64_encoded_32_byte_key  # Required
DB_MAX_OPEN_CONNS=0    # 0 = unlimited
DB_MAX_IDLE_CONNS=0    # 0 = driver default
DB_CONN_MAX_LIFETIME_SECS=0  # 0 = unlimited

MAIL_KEY=your_mail_key
MAIL_DOMAIN=your_mail_domain
MAIL_ENDPOINT=https://api.mailgun.net
```

### Configuration File (config.json)

```json
{
  "server": {
    "address": ":8080",
    "domain": "ToolsOfWorship.com",
    "verificationEmailTemplatePath": "./templates/VerificationEmailTemplate.html",
    "corsAllowedOrigins": ["https://example.com", "https://www.example.com"],
    "requestTimeoutSecs": 30
  },
  "database": {
    "ssl": true,
    "host": "localhost",
    "port": 5432,
    "user": "user",
    "password": "password",
    "name": "ToW",
    "masterKey": "base64_encoded_32_byte_key",
    "maxOpenConns": 0,
    "maxIdleConns": 0,
    "connMaxLifetimeSecs": 0
  },
  "mail": {
    "key": "your_mail_key",
    "domain": "your_mail_domain",
    "endpoint": "https://api.mailgun.net"
  }
}
```

### NGINX Configuration

NGINX handles SSL termination and serves static files. The `/metrics` endpoint must **not** be exposed publicly — restrict it to your Prometheus scraper IP.

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

    # Restrict metrics to Prometheus scraper only — never expose publicly
    location /metrics {
        allow <prometheus-scraper-ip>;
        deny all;
        proxy_pass http://localhost:8080/metrics;
    }
}
```

## Deployment

### GitHub Actions

Pushes to `main` automatically build and deploy. Required repository secrets:

| Secret            | Description                        |
| ----------------- | ---------------------------------- |
| `SERVER_HOST`     | Server IP or hostname              |
| `SERVER_USER`     | SSH username                       |
| `SSH_PRIVATE_KEY` | Private SSH key                    |
| `TARGET_DIR`      | Deployment directory on the server |

### Systemd Service

A sample service file is provided in `tow-server.service`. Run under a dedicated non-privileged user.

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

Environment=LISTEN_ADDRESS=:8080
Environment=DOMAIN=toolsofworship.com
Environment=DB_HOST=localhost
Environment=DB_USER=tow_user
Environment=DB_NAME=ToW
# Use EnvironmentFile for secrets (MASTER_KEY, MAIL_KEY, DB_PASSWORD)
# EnvironmentFile=/usr/local/lib/tools-of-worship/.env

NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=full
ProtectHome=yes

[Install]
WantedBy=multi-user.target
```

To install:

```bash
sudo useradd -r -s /bin/false tow
sudo mkdir -p /usr/local/lib/tools-of-worship
sudo cp tow-server /usr/local/lib/tools-of-worship/
sudo chown -R tow:tow /usr/local/lib/tools-of-worship
sudo cp tow-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable tow-server
sudo systemctl start tow-server
```

## Project Structure

```
├── api/                        # OpenAPI 3.0 specification
├── cmd/
│   └── tow-server/             # Entry point and configuration
├── internal/
│   ├── api/                    # HTTP server, middleware, routing, handlers
│   ├── cache/                  # In-process TTL cache wrappers
│   ├── config/                 # Configuration interfaces
│   ├── db/
│   │   └── postgresql/         # PostgreSQL store implementations
│   │       └── migrations/     # SQL migration files (embedded at compile time)
│   ├── domain/                 # Models, store interfaces, constants, errors
│   ├── keys/                   # Cryptographic operations
│   └── service/                # Business logic
├── public/                     # Static files served by nginx
└── templates/                  # Email templates
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
