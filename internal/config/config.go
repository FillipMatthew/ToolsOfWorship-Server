package config

type ServerConfig interface {
	IsTLS() bool
	GetListenAddress() string
	GetCertPath() string
	GetKeyPath() string
	GetPublicDir() string
	GetDomain() string
}

type KeysConfig interface {
	GetEncryptionKey() []byte // Must be 32 bytes for AES-256
	GetSigningKey() []byte    // HMAC signing secret
}

type DatabaseConfig interface {
	UseSSL() bool
	GetHost() string
	GetPort() uint
	GetUser() string
	GetPassword() string
	GetName() string
}
