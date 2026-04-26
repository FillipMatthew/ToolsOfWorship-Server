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

type MailConfig interface {
	GetMailKey() string
	GetMailDomain() string
	GetMailEndpoint() string
}
