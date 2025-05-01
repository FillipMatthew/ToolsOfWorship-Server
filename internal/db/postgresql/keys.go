package postgresql

import (
	"context"
	"database/sql"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
)

func NewKeyStore(db *sql.DB) *KeyStore {
	return &KeyStore{db: db}
}

type KeyStore struct {
	db *sql.DB
}

func (ks *KeyStore) GetSigningKey(ctx context.Context, id uuid.UUID) (domain.Key, error) {
	key := domain.Key{Id: id}

	err := ks.db.QueryRowContext(ctx, "SELECT key, expiry FROM SignKeys WHERE id=$1", id).Scan(&key.Key, &key.Expiry)
	if err != nil {
		return domain.Key{}, err
	}

	return key, nil
}

func (ks *KeyStore) GetEncryptionKey(ctx context.Context, id uuid.UUID) (domain.Key, error) {
	key := domain.Key{Id: id}

	err := ks.db.QueryRowContext(ctx, "SELECT key, expiry FROM EncKeys WHERE id=$1", id).Scan(&key.Key, &key.Expiry)
	if err != nil {
		return domain.Key{}, err
	}

	return key, nil
}

func (ks *KeyStore) GetSigningKeys(ctx context.Context) (map[uuid.UUID]domain.Key, error) {
	rows, err := ks.db.QueryContext(ctx, "SELECT (id, key, expiry) FROM SignKeys")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make(map[uuid.UUID]domain.Key)

	for rows.Next() {
		var key domain.Key
		err := rows.Scan(key.Id, key.Key, key.Expiry)
		if err != nil {
			return nil, err
		}

		keys[key.Id] = key
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func (ks *KeyStore) GetEncryptionKeys(ctx context.Context) (map[uuid.UUID]domain.Key, error) {
	rows, err := ks.db.QueryContext(ctx, "SELECT (id, key, expiry) FROM EncKeys")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make(map[uuid.UUID]domain.Key)

	for rows.Next() {
		var key domain.Key
		err := rows.Scan(key.Id, key.Key, key.Expiry)
		if err != nil {
			return nil, err
		}

		keys[key.Id] = key
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func (ks *KeyStore) SaveSigningKey(ctx context.Context, key domain.Key) error {
	_, err := ks.db.ExecContext(ctx, "INSERT INTO SignKeys (id, key, expiry) VALUES ($1, $2, $3)", key.Id, key.Key, key.Expiry)
	if err != nil {
		return err
	}

	return nil
}

func (ks *KeyStore) SaveEncryptionKey(ctx context.Context, key domain.Key) error {
	_, err := ks.db.ExecContext(ctx, "INSERT INTO EncKeys (id, key, expiry) VALUES ($1, $2, $3)", key.Id, key.Key, key.Expiry)
	if err != nil {
		return err
	}

	return nil
}

func (ks *KeyStore) RemoveSigningKey(ctx context.Context, id uuid.UUID) error {
	_, err := ks.db.ExecContext(ctx, "DELETE FROM SignKeys WHERE id=$1", id)
	if err != nil {
		return err
	}

	return nil
}

func (ks *KeyStore) RemoveEncryptionKey(ctx context.Context, id uuid.UUID) error {
	_, err := ks.db.ExecContext(ctx, "DELETE FROM EncKeys WHERE id=$1", id)
	if err != nil {
		return err
	}

	return nil
}
