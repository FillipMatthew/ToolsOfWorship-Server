package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func NewUserService(store domain.UserStore, tokensService TokensService, mailService MailService) *UserService {
	return &UserService{userStore: store, tokensService: tokensService, mailService: mailService}
}

type UserService struct {
	userStore     domain.UserStore
	tokensService TokensService
	mailService   MailService
}

func (u *UserService) Login(ctx context.Context, accountId, password string) (*domain.Token, *domain.User, error) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(accountId) {
		return nil, nil, errors.New("invalid email format")
	}

	accountId = strings.ToLower(accountId)

	userConnection, err := u.userStore.GetUserConnection(ctx, domain.LocalUser, accountId)
	if err != nil {
		return nil, nil, errors.New("login failed, invalid credentials")
	}

	bytePassword := []byte(password)
	byteHash := []byte(*userConnection.AuthDetails)
	err = bcrypt.CompareHashAndPassword(byteHash, bytePassword)
	if err != nil {
		return nil, nil, errors.New("login failed, invalid credentials")
	}

	user, err := u.userStore.GetUser(ctx, userConnection.UserId)
	if err != nil {
		return nil, nil, errors.New("unable to fetch user")
	}

	token, err := u.generateUserAuthToken(ctx, *user)
	if err != nil {
		return nil, nil, errors.New("unable to generate auth token")
	}

	return token, user, nil
}

func (u *UserService) SignIn(ctx context.Context, userConnection domain.UserConnection) (*domain.Token, *domain.User, error) {
	if userConnection.SignInType == domain.LocalUser && userConnection.AuthDetails != nil {
		return u.Login(ctx, userConnection.AccountId, *userConnection.AuthDetails)
	} else {
		return nil, nil, errors.New("invalid signin")
	}
}

func (u *UserService) Register(ctx context.Context, user domain.User, accountId, password string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(accountId) {
		return errors.New("invalid email format")
	}

	accountId = strings.ToLower(accountId)

	displayNameRegex := regexp.MustCompile(`^[a-zA-Z0-9 ]{3,30}$`)
	if !displayNameRegex.MatchString(user.DisplayName) {
		return errors.New("invalid display name")
	}

	if len(password) < 8 {
		return errors.New("password too short")
	}

	if !u.validateNewUser(ctx, accountId) {
		return errors.New("email already in use")
	}

	bytePassword := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return errors.New("internal error")
	}

	err = u.sendVerificationMail(ctx, accountId, string(hashedPassword), user.DisplayName)
	if err != nil {
		return errors.New("unable to send verification email")
	}

	return nil
}

func (u *UserService) ValidateUser(ctx context.Context, token domain.Token) (*domain.User, error) {
	userId, err := u.validateUserAuthToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user auth token: %v", err)
	}

	user, err := u.userStore.GetUser(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %v", err)
	}

	return user, nil
}

type newAccountDetails struct {
	Email       string `json:"email"`
	AuthDetails string `json:"authDetails"`
	DisplayName string `json:"displayName"`
}

func (u *UserService) VerifyAccount(ctx context.Context, token domain.Token) error {
	data, err := u.tokensService.VerifyEncryptedToken(ctx, string(token), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to verify token: %v", err)
	}

	sub, ok := data["sub"].(string)
	if !ok {
		return errors.New("invalid token data[sub]")
	}

	verificationDetails := newAccountDetails{}

	err = json.Unmarshal([]byte(sub), &verificationDetails)
	if err != nil {
		return errors.New("invalid token data")
	}

	_, err = u.createNewUser(ctx, verificationDetails.Email, verificationDetails.AuthDetails, verificationDetails.DisplayName)
	if err != nil {
		return fmt.Errorf("failed to create new user: %v", err)
	}

	return nil
}

func (u *UserService) validateNewUser(ctx context.Context, email string) bool {
	userConnection, err := u.userStore.GetUserConnection(ctx, domain.LocalUser, email)
	if err != nil {
		return true
	}

	return userConnection.IsValid()
}

func (u *UserService) createNewUser(ctx context.Context, email, authDetails, displayName string) (uuid.UUID, error) {
	userId, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, errors.New("could not generate an id")
	}

	err = u.userStore.CreateUser(ctx, domain.User{Id: userId, DisplayName: displayName})
	if err != nil {
		return uuid.Nil, errors.New("could not save user")
	}

	userConnection := domain.UserConnection{UserId: userId, SignInType: domain.LocalUser, AccountId: email, AuthDetails: &authDetails}
	err = u.userStore.SaveUserConnection(ctx, userConnection)
	if err != nil {
		u.userStore.RemoveUser(ctx, userId)
		return uuid.Nil, errors.New("could not save user connection")
	}

	return userId, nil
}

func (u *UserService) sendVerificationMail(ctx context.Context, email, authDetails, displayName string) error {
	verificationDetails := newAccountDetails{
		Email:       email,
		AuthDetails: authDetails,
		DisplayName: displayName,
	}

	jsonData, err := json.Marshal(verificationDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal email verification details: %v", err)
	}

	payload := map[string]any{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"sub": string(jsonData),
	}

	token, err := u.tokensService.SignEncryptedToken(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to sign token: %v", err)
	}

	templatePath := "./templates/VerificationEmailTemplate.html"
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %v", err)
	}

	contentStr := strings.ReplaceAll(string(content), "@token", token)

	err = u.mailService.SendNoReplyEmail(displayName, email, "Please verify your email address", contentStr)
	if err != nil {
		return err
	}

	return nil
}

type userAuthToken struct {
	UserId uuid.UUID `json:"userId"`
}

func (u *UserService) generateUserAuthToken(ctx context.Context, user domain.User) (*domain.Token, error) {
	tokenDetails := userAuthToken{UserId: user.Id}

	jsonData, err := json.Marshal(tokenDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user details: %v", err)
	}

	payload := map[string]any{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"sub": string(jsonData),
	}

	token, err := u.tokensService.SignJWT(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %v", err)
	}

	result := domain.Token(token)
	return &result, nil
}

func (u *UserService) validateUserAuthToken(ctx context.Context, token domain.Token) (uuid.UUID, error) {
	data, err := u.tokensService.VerifyJWT(string(token))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to verify token: %v", err)
	}

	sub, ok := data["sub"].(string)
	if !ok {
		return uuid.UUID{}, errors.New("invalid token data[sub]")
	}

	tokenDetails := userAuthToken{}

	err = json.Unmarshal([]byte(sub), &tokenDetails)
	if err != nil {
		return uuid.UUID{}, errors.New("invalid token data")
	}

	return tokenDetails.UserId, nil
}
