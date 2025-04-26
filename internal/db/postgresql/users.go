package postgresql

import (
	"context"
	"database/sql"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
)

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) GetUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user := domain.User{Id: id}

	err := u.db.QueryRow("SELECT displayName FROM Users WHERE id=$1", id).Scan(user.DisplayName)
	if err != nil {
		//log.Fatalln(err)
		return domain.User{}, err
	}

	return user, nil
}

func (u *UserStore) GetUserConnection(ctx context.Context, signInType domain.SignInType, accountId string) (domain.UserConnection, error) {
	conn := domain.UserConnection{SignInType: signInType, AccountId: accountId}

	err := u.db.QueryRow("SELECT userId, authDetails FROM UserConnections WHERE signInType=$1 AND accountId=$2", signInType, accountId).Scan(conn.UserId, conn.AuthDetails)
	if err != nil {
		//log.Fatalln(err)
		return domain.UserConnection{}, err
	}

	return conn, nil
}

func (u *UserStore) CreateUser(ctx context.Context, user domain.User) error {
	if user.Id == uuid.Nil {
		//log.Fatalln(err)
		panic("invalid user id")
	}

	_, err := u.db.Exec("INSERT INTO Users (id, displayName) VALUES ($1, $2)", user.Id, user.DisplayName)
	if err != nil {
		//log.Fatalln(err)
		return err
	}

	return nil
}

func (u *UserStore) RemoveUser(ctx context.Context, id uuid.UUID) error {
	_, err := u.db.Exec("DELETE FROM UserConnections WHERE userId=$1;DELETE FROM Users WHERE id=$1;", id)
	if err != nil {
		//log.Fatalln(err)
		return err
	}

	return nil
}

func (u *UserStore) SaveUserConnection(ctx context.Context, userConnection domain.UserConnection) error {
	if userConnection.UserId == uuid.Nil {
		//log.Fatalln(err)
		panic("invalid userId on user connection")
	}

	_, err := u.db.Exec("INSERT INTO UserConnections (userId, signInType, accountId, authDetails) VALUES ($1, $2, $3, $4)", userConnection.UserId, userConnection.SignInType, userConnection.AccountId, userConnection.AuthDetails)
	if err != nil {
		//log.Fatalln(err)
		return err
	}

	return nil
}
