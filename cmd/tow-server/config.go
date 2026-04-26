package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type serverConfig struct {
	ListenAddress                 string   `json:"address"`
	Domain                        string   `json:"domain"`
	VerificationEmailTemplatePath string   `json:"verificationEmailTemplatePath"`
	CORSAllowedOrigins            []string `json:"corsAllowedOrigins"`
	RequestTimeoutSecs            int      `json:"requestTimeoutSecs"`
}

type databaseConfig struct {
	UseSSL              bool   `json:"ssl"`
	Host                string `json:"host"`
	Port                uint   `json:"port"`
	User                string `json:"user"`
	Password            string `json:"password"`
	Name                string `json:"name"`
	MasterKey           []byte `json:"masterKey"`
	MaxOpenConns        int    `json:"maxOpenConns"`
	MaxIdleConns        int    `json:"maxIdleConns"`
	ConnMaxLifetimeSecs int    `json:"connMaxLifetimeSecs"`
}

type mailConfig struct {
	Key      string `json:"key"`
	Domain   string `json:"domain"`
	Endpoint string `json:"endpoint"`
}

type config struct {
	Server   serverConfig   `json:"server"`
	Database databaseConfig `json:"database"`
	Mail     mailConfig     `json:"mail"`
}

func (config *config) GetListenAddress() string {
	return config.Server.ListenAddress
}

func (config *config) GetDomain() string {
	return config.Server.Domain
}

func (config *config) GetVerificationEmailTemplatePath() string {
	return config.Server.VerificationEmailTemplatePath
}

func (config *config) GetCORSAllowedOrigins() []string {
	return config.Server.CORSAllowedOrigins
}

func (config *config) GetRequestTimeout() time.Duration {
	if config.Server.RequestTimeoutSecs <= 0 {
		return 30 * time.Second
	}
	return time.Duration(config.Server.RequestTimeoutSecs) * time.Second
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

func (config *config) GetMaxOpenConns() int { return config.Database.MaxOpenConns }
func (config *config) GetMaxIdleConns() int { return config.Database.MaxIdleConns }
func (config *config) GetConnMaxLifetime() time.Duration {
	if config.Database.ConnMaxLifetimeSecs <= 0 {
		return 0
	}
	return time.Duration(config.Database.ConnMaxLifetimeSecs) * time.Second
}

func (config *config) GetMailKey() string {
	return config.Mail.Key
}

func (config *config) GetMailDomain() string {
	return config.Mail.Domain
}

func (config *config) GetMailEndpoint() string {
	return config.Mail.Endpoint
}

func (c *config) Validate() error {
	if c.Server.Domain == "" {
		return fmt.Errorf("server domain is required")
	}
	if len(c.Database.MasterKey) != 32 {
		return fmt.Errorf("database master key must be 32 bytes")
	}
	if c.Mail.Key == "" {
		return fmt.Errorf("mail key is required")
	}
	if c.Mail.Domain == "" {
		return fmt.Errorf("mail domain is required")
	}
	if c.Mail.Endpoint == "" {
		return fmt.Errorf("mail endpoint is required")
	}
	return nil
}

func getConfig() *config {
	config := getEnvConfig()

	// Overwrite defaults and env with config file
	err := config.loadConfig("config.json")
	if err != nil {
		fmt.Println("Error loading config from json:", err)
	}

	// Overwrite all with any manually specified options
	flag.StringVar(&config.Server.ListenAddress, "address", config.Server.ListenAddress, "[Address:Port] to listen on")
	flag.StringVar(&config.Server.Domain, "domain", config.Server.Domain, "The base domain for the server endpoints (example.com)")
	flag.StringVar(&config.Server.VerificationEmailTemplatePath, "verificationEmailTemplatePath", config.Server.VerificationEmailTemplatePath, "Path to the verification email template")
	flag.IntVar(&config.Server.RequestTimeoutSecs, "requestTimeoutSecs", config.Server.RequestTimeoutSecs, "HTTP request timeout in seconds (default 30)")

	corsOrigins := flag.String("corsAllowedOrigins", "", "Comma-separated list of allowed CORS origins (empty = allow all)")

	flag.BoolVar(&config.Database.UseSSL, "dbssl", config.Database.UseSSL, "Use SSL for database?")
	flag.StringVar(&config.Database.Host, "dbhost", config.Database.Host, "Database host")
	flag.UintVar(&config.Database.Port, "dbport", config.Database.Port, "Database port")
	flag.StringVar(&config.Database.User, "dbuser", config.Database.User, "Database user")
	flag.StringVar(&config.Database.Password, "dbpassword", config.Database.Password, "Database password")
	flag.StringVar(&config.Database.Name, "dbname", config.Database.Name, "Database name")

	masterKey := flag.String("masterKey", "", "The master key used for encrypting keys in the DB (base64-encoded 32 bytes)")

	flag.IntVar(&config.Database.MaxOpenConns, "dbMaxOpenConns", config.Database.MaxOpenConns, "Max open DB connections (0 = unlimited)")
	flag.IntVar(&config.Database.MaxIdleConns, "dbMaxIdleConns", config.Database.MaxIdleConns, "Max idle DB connections (0 = driver default)")
	flag.IntVar(&config.Database.ConnMaxLifetimeSecs, "dbConnMaxLifetimeSecs", config.Database.ConnMaxLifetimeSecs, "Max DB connection lifetime in seconds (0 = unlimited)")

	flag.StringVar(&config.Mail.Key, "mailkey", config.Mail.Key, "Mail API key")
	flag.StringVar(&config.Mail.Domain, "maildomain", config.Mail.Domain, "Mail domain")
	flag.StringVar(&config.Mail.Endpoint, "mailendpoint", config.Mail.Endpoint, "Mail API endpoint")

	flag.Parse()

	// Apply flag overrides that need post-processing
	if len(*corsOrigins) != 0 {
		config.Server.CORSAllowedOrigins = splitTrimmed(*corsOrigins, ",")
	}

	if len(*masterKey) != 0 {
		keyBytes, err := base64.RawURLEncoding.DecodeString(*masterKey)
		if err != nil {
			fmt.Println("Error loading master key from command line args:", err)
		} else {
			if len(keyBytes) != 32 {
				fmt.Println("Error parsing master key: key must be exactly 32 bytes encoded in base64")
			} else {
				config.Database.MasterKey = keyBytes
			}
		}
	}

	if err := config.Validate(); err != nil {
		fmt.Printf("Config validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config: %+v\n", config)

	return config
}

func getEnvConfig() *config {
	listenAddress := os.Getenv("LISTEN_ADDRESS")
	if listenAddress == "" {
		listenAddress = ":8080"
	}

	domain := os.Getenv("DOMAIN")

	verificationEmailTemplatePath := os.Getenv("VERIFICATION_EMAIL_TEMPLATE_PATH")
	if verificationEmailTemplatePath == "" {
		verificationEmailTemplatePath = "./templates/VerificationEmailTemplate.html"
	}

	var corsAllowedOrigins []string
	if raw := os.Getenv("CORS_ALLOWED_ORIGINS"); raw != "" {
		corsAllowedOrigins = splitTrimmed(raw, ",")
	}

	requestTimeoutSecs, err := strconv.Atoi(os.Getenv("REQUEST_TIMEOUT_SECS"))
	if err != nil || requestTimeoutSecs <= 0 {
		requestTimeoutSecs = 0 // zero triggers the default in GetRequestTimeout
	}

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

	dbport := uint(val)

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

	var masterKey []byte
	if masterKeyStr := os.Getenv("MASTER_KEY"); masterKeyStr != "" {
		masterKey, err = base64.RawURLEncoding.DecodeString(masterKeyStr)
		if err != nil {
			fmt.Println("Error loading master key from env:", err)
		} else if len(masterKey) != 32 {
			fmt.Println("Error parsing master key from env: key must be exactly 32 bytes encoded in base64")
			masterKey = nil
		}
	}

	dbMaxOpenConns, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	if err != nil {
		dbMaxOpenConns = 0
	}

	dbMaxIdleConns, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	if err != nil {
		dbMaxIdleConns = 0
	}

	dbConnMaxLifetimeSecs, err := strconv.Atoi(os.Getenv("DB_CONN_MAX_LIFETIME_SECS"))
	if err != nil {
		dbConnMaxLifetimeSecs = 0
	}

	mailkey := os.Getenv("MAIL_KEY")
	maildomain := os.Getenv("MAIL_DOMAIN")
	mailendpoint := os.Getenv("MAIL_ENDPOINT")

	return &config{
		Server: serverConfig{
			ListenAddress:                 listenAddress,
			Domain:                        domain,
			VerificationEmailTemplatePath: verificationEmailTemplatePath,
			CORSAllowedOrigins:            corsAllowedOrigins,
			RequestTimeoutSecs:            requestTimeoutSecs,
		},
		Database: databaseConfig{
			UseSSL:              useSSL,
			Host:                dbhost,
			Port:                dbport,
			User:                dbuser,
			Password:            dbpassword,
			Name:                dbname,
			MasterKey:           masterKey,
			MaxOpenConns:        dbMaxOpenConns,
			MaxIdleConns:        dbMaxIdleConns,
			ConnMaxLifetimeSecs: dbConnMaxLifetimeSecs,
		},
		Mail: mailConfig{
			Key:      mailkey,
			Domain:   maildomain,
			Endpoint: mailendpoint,
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

// splitTrimmed splits s by sep and trims whitespace from each element,
// omitting any empty strings that result.
func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
