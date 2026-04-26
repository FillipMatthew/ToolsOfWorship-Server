package config

import "time"

type ServerConfig interface {
	GetListenAddress() string
	GetDomain() string
	GetVerificationEmailTemplatePath() string
	GetCORSAllowedOrigins() []string
	GetRequestTimeout() time.Duration
}

type DatabaseConfig interface {
	UseSSL() bool
	GetHost() string
	GetPort() uint
	GetUser() string
	GetPassword() string
	GetName() string
	GetMasterKey() []byte
}

// DatabasePoolConfig configures the database connection pool.
// Zero values mean "use driver default".
type DatabasePoolConfig interface {
	GetMaxOpenConns() int
	GetMaxIdleConns() int
	GetConnMaxLifetime() time.Duration
}

type MailConfig interface {
	GetMailKey() string
	GetMailDomain() string
	GetMailEndpoint() string
}
