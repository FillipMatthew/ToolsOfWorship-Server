package config

type ServerConfig interface {
	IsTLS() bool
	GetListenAddress() string
	GetCertPath() string
	GetKeyPath() string
	GetPublicDir() string
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
