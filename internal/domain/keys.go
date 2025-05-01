package domain

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/google/uuid"
)

type Key struct {
	Id     uuid.UUID
	Key    []byte
	Expiry time.Time
}

type KeyStoreReader interface {
	GetSigningKey(ctx context.Context, id uuid.UUID) (Key, error)
	GetEncryptionKey(ctx context.Context, id uuid.UUID) (Key, error)
	GetSigningKeys(ctx context.Context) (map[uuid.UUID]Key, error)
	GetEncryptionKeys(ctx context.Context) (map[uuid.UUID]Key, error)
}

type KeyStoreWriter interface {
	SaveSigningKey(ctx context.Context, key Key) error
	SaveEncryptionKey(ctx context.Context, key Key) error
	RemoveSigningKey(ctx context.Context, id uuid.UUID) error
	RemoveEncryptionKey(ctx context.Context, id uuid.UUID) error
}

type KeyStore interface {
	KeyStoreReader
	KeyStoreWriter
}

func (k *Key) IsValid() bool {
	if k.Id == uuid.Nil {
		return false
	}

	if len(k.Key) == 0 {
		return false
	}

	if k.Expiry.Unix() < time.Now().Unix() {
		return false
	}

	return true
}

func NewKey() (Key, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return Key{}, nil
	}

	newKey := Key{
		Id:     uuid.New(),
		Key:    key,
		Expiry: time.Now().Add(182 * 24 * time.Hour),
	}

	return newKey, nil
}
