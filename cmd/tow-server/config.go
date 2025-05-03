package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type serverConfig struct {
	IsTLS         bool   `json:"tls"`
	ListenAddress string `json:"address"`
	CertPath      string `json:"cert"`
	KeyPath       string `json:"key"`
	PublicDir     string `json:"public"`
	Domain        string `json:"domain"`
}

type databaseConfig struct {
	UseSSL    bool   `json:"ssl"`
	Host      string `json:"host"`
	Port      uint   `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	MasterKey []byte `json:"masterKey"`
}

type config struct {
	Server   serverConfig   `json:"server"`
	Database databaseConfig `json:"database"`
}

func (config *config) IsTLS() bool {
	return config.Server.IsTLS
}

func (config *config) GetListenAddress() string {
	return config.Server.ListenAddress
}

func (config *config) GetCertPath() string {
	return config.Server.CertPath
}

func (config *config) GetKeyPath() string {
	return config.Server.KeyPath
}

func (config *config) GetPublicDir() string {
	return config.Server.PublicDir
}

func (config *config) GetDomain() string {
	return config.Server.Domain
}

func (config *config) UseSSL() bool {
	return config.Database.UseSSL
}

func (config *config) GetHost() string {
	return config.Database.Host
}

func (config *config) GetPort() uint {
	return config.Database.Port
}

func (config *config) GetUser() string {
	return config.Database.User
}

func (config *config) GetPassword() string {
	return config.Database.Password
}

func (config *config) GetName() string {
	return config.Database.Name
}

func (config *config) GetMasterKey() []byte {
	return config.Database.MasterKey
}

func getConfig() *config {
	config := getEnvConfig()

	// Overwrite defaults and env with config file
	err := config.loadConfig("config.json")
	if err != nil {
		fmt.Println("Error loading config from json:", err)
	}

	// Overwrite all with any manually specified options
	flag.BoolVar(&config.Server.IsTLS, "tls", config.Server.IsTLS, "Use TLS?")
	flag.StringVar(&config.Server.ListenAddress, "address", config.Server.ListenAddress, "[Address:Port] to listen on")
	flag.StringVar(&config.Server.CertPath, "cert", config.Server.CertPath, "TLS certificate path")
	flag.StringVar(&config.Server.KeyPath, "key", config.Server.KeyPath, "TLS private key path")
	flag.StringVar(&config.Server.PublicDir, "public", config.Server.PublicDir, "Public directory path")
	flag.StringVar(&config.Server.Domain, "domain", config.Server.Domain, "The base domain for the server endpoints (example.com)")
	flag.BoolVar(&config.Database.UseSSL, "dbssl", config.Database.UseSSL, "Use SSL for database?")
	flag.StringVar(&config.Database.Host, "dbhost", config.Database.Host, "Database host")
	flag.UintVar(&config.Database.Port, "dbport", config.Database.Port, "Database port")
	flag.StringVar(&config.Database.User, "dbuser", config.Database.User, "Database user")
	flag.StringVar(&config.Database.Password, "dbpassword", config.Database.Password, "Database password")
	flag.StringVar(&config.Database.Name, "dbname", config.Database.Name, "Database name")
	masterKey := flag.String("masterKey", "", "The master used for encrypting keys in the DB.")
	if len(*masterKey) != 0 {
		keyBytes, err := base64.RawURLEncoding.DecodeString(*masterKey)
		if err != nil {
			fmt.Println("Error loading master key command line args:", err)
		} else {
			config.Database.MasterKey = keyBytes
		}

		if len(keyBytes) != 32 {
			fmt.Print("Error parsing master key. Key must be 32 bytes encoded in base64")
		}
	}

	flag.Parse()

	fmt.Printf("Config: %+v\n", config)

	return config
}

func getEnvConfig() *config {
	isTLS, err := strconv.ParseBool(os.Getenv("USE_TLS"))
	if err != nil {
		isTLS = true // Default value
	}

	listenAddress := os.Getenv("LISTEN_ADDRESS")
	if listenAddress == "" {
		listenAddress = ":443"
	}

	certPath := os.Getenv("CERT_PATH")
	if certPath == "" {
		certPath = "./certs/cert.pem"
	}

	keyPath := os.Getenv("CERT_PATH")
	if keyPath == "" {
		keyPath = "./certs/key.pem"
	}

	publicDir := os.Getenv("PUBLIC_DIR")

	domain := os.Getenv("DOMAIN")

	useSSL, err := strconv.ParseBool(os.Getenv("DB_USE_SSL"))
	if err != nil {
		useSSL = true // Default value
	}

	dbhost := os.Getenv("DB_HOST")
	if dbhost == "" {
		dbhost = "localhost"
	}

	val, err := strconv.ParseUint(os.Getenv("DB_PORT"), 10, 32)
	if err != nil {
		val = 5432
	}

	var dbport uint = uint(val)

	dbuser := os.Getenv("DB_USER")
	if dbuser == "" {
		dbuser = "user"
	}

	dbpassword := os.Getenv("DB_PASSWORD")
	if dbpassword == "" {
		dbpassword = "password"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "ToW"
	}

	masterKeyStr := os.Getenv("MASTER_KEY")
	var masterKey []byte
	if len(masterKeyStr) != 0 {
		masterKey, err := base64.RawURLEncoding.DecodeString(masterKeyStr)
		if err != nil {
			fmt.Println("Error loading master key command line args:", err)
		}

		if len(masterKey) != 32 {
			fmt.Print("Error parsing master key. Key must be 32 bytes encoded in base64")
		}
	}

	return &config{
		Server: serverConfig{
			IsTLS:         isTLS,
			ListenAddress: listenAddress,
			CertPath:      certPath,
			KeyPath:       keyPath,
			PublicDir:     publicDir,
			Domain:        domain,
		},
		Database: databaseConfig{
			UseSSL:    useSSL,
			Host:      dbhost,
			Port:      uint(dbport),
			User:      dbuser,
			Password:  dbpassword,
			Name:      dbname,
			MasterKey: masterKey,
		},
	}
}

func (c *config) loadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		return err
	}

	return nil
}
