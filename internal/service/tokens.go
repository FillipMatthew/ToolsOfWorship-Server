package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/config"
)

func NewTokensService(config config.ServerConfig) *TokensService {
	return &TokensService{config: config}
}

type TokensService struct {
	config config.ServerConfig
}

func (ts *TokensService) SignJWT(payload map[string]any, signingKey []byte) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	payload["iss"] = "https://" + ts.config.GetDomain()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := headerB64 + "." + payloadB64
	signature, err := sign([]byte(signingInput), signingKey)
	if err != nil {
		return "", err
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	token := signingInput + "." + signatureB64
	return token, nil
}

func (ts *TokensService) VerifyJWT(token string, signingKey []byte) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	signingInput := headerB64 + "." + payloadB64
	signature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, err
	}

	if !verifySignature([]byte(signingInput), signature, signingKey) {
		return nil, errors.New("invalid signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	// Check expiration
	if expValue, ok := payload["exp"]; ok {
		switch exp := expValue.(type) {
		case float64:
			if int64(exp) < time.Now().Unix() {
				return nil, errors.New("token has expired")
			}
		case int64:
			if exp < time.Now().Unix() {
				return nil, errors.New("token has expired")
			}
		default:
			return nil, errors.New("invalid exp field format")
		}
	}

	return payload, nil
}

func (ts *TokensService) SignEncryptedToken(payload map[string]any, encryptionKey, signingKey []byte) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"enc": "A256GCM",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	encryptedPayload, err := encryptAESGCM(payloadJSON, encryptionKey)
	if err != nil {
		return "", err
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(encryptedPayload)

	signingInput := headerB64 + "." + payloadB64
	signature, err := sign([]byte(signingInput), signingKey)
	if err != nil {
		return "", err
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	token := signingInput + "." + signatureB64
	return token, nil
}

func (ts *TokensService) VerifyEncryptedToken(token string, encryptionKey, signingKey []byte) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	signingInput := headerB64 + "." + payloadB64
	signature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, err
	}

	if !verifySignature([]byte(signingInput), signature, signingKey) {
		return nil, errors.New("invalid signature")
	}

	encryptedPayload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, err
	}

	decryptedPayload, err := decryptAESGCM(encryptedPayload, encryptionKey)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(decryptedPayload, &payload); err != nil {
		return nil, err
	}

	// Expiration checking
	if expValue, ok := payload["exp"]; ok {
		switch exp := expValue.(type) {
		case float64:
			if int64(exp) < time.Now().Unix() {
				return nil, errors.New("token has expired")
			}
		case int64:
			if exp < time.Now().Unix() {
				return nil, errors.New("token has expired")
			}
		default:
			return nil, errors.New("invalid exp field format")
		}
	}

	return payload, nil
}

func sign(data, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(data)
	if err != nil {
		return nil, err
	}

	return mac.Sum(nil), nil
}

func verifySignature(data, signature, key []byte) bool {
	expected, err := sign(data, key)
	if err != nil {
		return false
	}

	return hmac.Equal(expected, signature)
}

func encryptAESGCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func decryptAESGCM(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, cipherdata := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesgcm.Open(nil, nonce, cipherdata, nil)
}
