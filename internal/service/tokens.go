package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/config"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/keys"
	"github.com/google/uuid"
)

func NewTokensService(ctx context.Context, config config.ServerConfig, keyStore domain.KeyStore) *TokensService {
	tokensService := &TokensService{config: config, keyStore: keyStore}
	err := tokensService.initialise(ctx)
	if err != nil {
		fmt.Printf("failed to initialise tokens service: %v\n", err)
		return nil
	}

	return tokensService
}

type TokensService struct {
	config                config.ServerConfig
	keyStore              domain.KeyStore
	currentSigningKey     domain.Key
	signingKeys           map[uuid.UUID]domain.Key
	currentEncryptionKey  domain.Key
	previousEncryptionKey domain.Key
}

func (ts *TokensService) initialise(ctx context.Context) error {
	signingKeys, err := ts.keyStore.GetSigningKeys(ctx)
	if err != nil {
		return fmt.Errorf("could not get signing keys: %v", err)
	}

	ts.signingKeys = signingKeys
	for _, key := range ts.signingKeys {
		if ts.currentSigningKey.Expiry.IsZero() || (!key.Expiry.IsZero() && ts.currentSigningKey.Expiry.Unix() < key.Expiry.Unix()) {
			ts.currentSigningKey = key
		}
	}

	encryptionKeys, err := ts.keyStore.GetEncryptionKeys(ctx)
	if err != nil {
		return fmt.Errorf("could not get encryption keys: %v", err)
	}

	for _, key := range encryptionKeys {
		if ts.currentEncryptionKey.Expiry.IsZero() || (!key.Expiry.IsZero() && ts.currentSigningKey.Expiry.Unix() < key.Expiry.Unix()) {
			ts.previousEncryptionKey = ts.currentEncryptionKey
			ts.currentEncryptionKey = key
		} else if ts.previousEncryptionKey.Expiry.IsZero() || (!key.Expiry.IsZero() && ts.previousEncryptionKey.Expiry.Unix() < key.Expiry.Unix()) {
			ts.previousEncryptionKey = key // If the first item is the latest key then the previous would not get set without this
		}
	}

	return nil
}

func (ts *TokensService) SignJWT(ctx context.Context, payload map[string]any) (string, error) {
	return ts.SignJWTWithKey(ctx, payload, nil)
}

func (ts *TokensService) VerifyJWT(token string) (map[string]any, error) {
	return ts.VerifyJWTWithKey(token, nil)
}

func (ts *TokensService) SignEncryptedToken(ctx context.Context, payload map[string]any) (string, error) {
	return ts.SignEncryptedTokenWithKey(ctx, payload, nil, nil)
}

func (ts *TokensService) VerifyEncryptedToken(ctx context.Context, token string, encryptionKey, signingKey []byte) (map[string]any, error) {
	return ts.VerifyEncryptedTokenWithKey(ctx, token, nil, nil)
}

func (ts *TokensService) SignJWTWithKey(ctx context.Context, payload map[string]any, signingKey []byte) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	if len(signingKey) == 0 {
		key, err := ts.getCurrentSigningKey(ctx)
		if err != nil {
			return "", err
		}

		header["kid"] = key.Id.String()
		signingKey = key.Key
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
	signature, err := keys.Sign([]byte(signingInput), signingKey)
	if err != nil {
		return "", err
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	token := signingInput + "." + signatureB64
	return token, nil
}

func (ts *TokensService) VerifyJWTWithKey(token string, signingKey []byte) (map[string]any, error) {
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

	if len(signingKey) == 0 {
		headerBytes, err := base64.RawURLEncoding.DecodeString(headerB64)
		if err != nil {
			return nil, err
		}

		var header map[string]any
		if err := json.Unmarshal(headerBytes, &header); err != nil {
			return nil, err
		}

		if kid, ok := header["kid"].(string); ok {
			id, err := uuid.Parse(kid)
			if err != nil {
				return nil, err
			}

			key, err := ts.getSigningKey(id)
			if err != nil {
				return nil, err
			}

			signingKey = key.Key
		} else {
			return nil, errors.New("invalid or missing 'kid'")
		}
	}

	if !keys.VerifySignature([]byte(signingInput), signature, signingKey) {
		return nil, errors.New("invalid signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, err
	}

	var payload map[string]any
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

func (ts *TokensService) SignEncryptedTokenWithKey(ctx context.Context, payload map[string]any, encryptionKey, signingKey []byte) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"enc": "A256GCM",
		"typ": "JWT",
	}

	if len(signingKey) == 0 {
		key, err := ts.getCurrentSigningKey(ctx)
		if err != nil {
			return "", err
		}

		header["kid"] = key.Id.String()
		signingKey = key.Key
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

	if len(encryptionKey) == 0 {
		key, err := ts.getCurrentEncryptionKey(ctx)
		if err != nil {
			return "", err
		}

		encryptionKey = key.Key
	}

	encryptedPayload, err := keys.EncryptAESGCM(payloadJSON, encryptionKey)
	if err != nil {
		return "", err
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(encryptedPayload)

	signingInput := headerB64 + "." + payloadB64
	signature, err := keys.Sign([]byte(signingInput), signingKey)
	if err != nil {
		return "", err
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	token := signingInput + "." + signatureB64
	return token, nil
}

func (ts *TokensService) VerifyEncryptedTokenWithKey(ctx context.Context, token string, encryptionKey, signingKey []byte) (map[string]any, error) {
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

	if len(signingKey) == 0 {
		headerBytes, err := base64.RawURLEncoding.DecodeString(headerB64)
		if err != nil {
			return nil, err
		}

		var header map[string]any
		if err := json.Unmarshal(headerBytes, &header); err != nil {
			return nil, err
		}

		if kid, ok := header["kid"].(string); ok {
			id, err := uuid.Parse(kid)
			if err != nil {
				return nil, err
			}

			key, err := ts.getSigningKey(id)
			if err != nil {
				return nil, err
			}

			signingKey = key.Key
		} else {
			return nil, errors.New("invalid or missing 'kid'")
		}
	}

	if !keys.VerifySignature([]byte(signingInput), signature, signingKey) {
		return nil, errors.New("invalid signature")
	}

	encryptedPayload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, err
	}

	if len(encryptionKey) == 0 {
		key, err := ts.getCurrentEncryptionKey(ctx)
		if err != nil {
			return nil, err
		}

		encryptionKey = key.Key
	}

	decryptedPayload, err := keys.DecryptAESGCM(encryptedPayload, encryptionKey)
	if err != nil {
		// Try previous encryption key before failing
		key, err2 := ts.getPreviousEncryptionKey()
		if err2 != nil {
			return nil, err // Return original error instead of the failure to get previous key
		}

		encryptionKey = key.Key
		decryptedPayload, err2 = keys.DecryptAESGCM(encryptedPayload, encryptionKey)
		if err2 != nil {
			return nil, err // Return original error instead of the failure to decrypt with the previous key
		}

		// Success on previous key, proceed as normal
	}

	var payload map[string]any
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

func (ts *TokensService) getCurrentSigningKey(ctx context.Context) (domain.Key, error) {
	if ts.currentSigningKey.IsValid() {
		return ts.currentSigningKey, nil
	}

	newKey, err := domain.NewKey()
	if err != nil {
		return domain.Key{}, err
	}

	if !newKey.IsValid() {
		return domain.Key{}, errors.New("invalid key generated")
	}

	ts.currentSigningKey = newKey
	if ts.signingKeys == nil {
		ts.signingKeys = map[uuid.UUID]domain.Key{}
	}

	ts.signingKeys[newKey.Id] = newKey

	err = ts.keyStore.SaveSigningKey(ctx, newKey)
	if err != nil {
		return domain.Key{}, err
	}

	return ts.currentSigningKey, nil
}

func (ts *TokensService) getSigningKey(id uuid.UUID) (domain.Key, error) {
	if key, exists := ts.signingKeys[id]; exists {
		return key, nil
	}

	return domain.Key{}, errors.New("key not found")
}

func (ts *TokensService) getCurrentEncryptionKey(ctx context.Context) (domain.Key, error) {
	if ts.currentEncryptionKey.IsValid() {
		return ts.currentEncryptionKey, nil
	}

	newKey, err := domain.NewKey()
	if err != nil {
		return domain.Key{}, err
	}

	if !newKey.IsValid() {
		return domain.Key{}, errors.New("invalid key generated")
	}

	if ts.previousEncryptionKey.IsValid() {
		ts.keyStore.RemoveEncryptionKey(ctx, ts.previousEncryptionKey.Id)
	}

	ts.previousEncryptionKey = ts.currentEncryptionKey
	ts.currentEncryptionKey = newKey
	err = ts.keyStore.SaveEncryptionKey(ctx, newKey)
	if err != nil {
		return domain.Key{}, err
	}

	return ts.currentEncryptionKey, nil
}

func (ts *TokensService) getPreviousEncryptionKey() (domain.Key, error) {
	if ts.previousEncryptionKey.Id != uuid.Nil && len(ts.previousEncryptionKey.Key) != 0 {
		return ts.previousEncryptionKey, nil
	}

	return domain.Key{}, errors.New("key not found")
}
