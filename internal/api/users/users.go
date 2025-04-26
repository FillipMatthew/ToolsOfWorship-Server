package users

import (
	"github.com/google/uuid"
)

type User struct {
	Id          uuid.UUID `json:"id"`
	DisplayName string    `json:"displayName"`
}

type RegisterUser struct {
	DisplayName string `json:"displayName"`
	AccountId   string `json:"accountId"`
	Password    string `json:"password"`
}

type RegisterResponse struct {
	ID uuid.UUID `json:"id"`
}
