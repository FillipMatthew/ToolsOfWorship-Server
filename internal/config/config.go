package config

type ServerConfig interface {
	GetListenAddress() string
	GetDomain() string
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
}
