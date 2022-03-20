package auth

import (
	"database/sql"
	"fmt"
	"github.com/Nethius/tribble-customer-auth/pkg/model"
	"github.com/Nethius/tribble-customer-auth/pkg/storage"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
)

type service struct {
	store         storage.Storage
	accessSecret  string
	refreshSecret string
}

func NewService(store storage.Storage) *service {
	s := service{store: store}
	s.setupTokensSecrets()

	return &s
}

func (s *service) setupTokensSecrets() {
	s.accessSecret = os.Getenv("ACCESSSECRET")
	s.refreshSecret = os.Getenv("REFRESHSECRET")
}

func (s *service) GenerateTokenPair(userID uint) (model.TokenPair, error) {

	if s.refreshSecret == "" || s.accessSecret == "" {
		return model.TokenPair{}, ErrMissingTokenSecret
	}

	accessToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"),
		&model.AccessToken{UserID: userID, ExpiresAt: time.Now().Add(time.Minute * 15).Unix()})

	refreshToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"),
		&model.RefreshToken{UserID: userID, ExpiresAt: time.Now().Add(time.Minute * 30).Unix()})

	accessTokenString, err := accessToken.SignedString([]byte(s.accessSecret))
	if err != nil {
		return model.TokenPair{}, err
	}
	refreshTokenString, err := refreshToken.SignedString([]byte(s.refreshSecret))
	if err != nil {
		return model.TokenPair{}, err
	}

	return model.TokenPair{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}

//func IsExists(email string) (string, error) {
//
//}
//
//func IsValid(account *model.Users) (string, error) {
//
//}

func (s *service) Register(email string, password string) (model.TokenPair, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.TokenPair{}, fmt.Errorf("failed to encrypt password: %w", err)
	}

	userID, err := s.store.InsertUser(email, string(hashedPassword))
	if err != nil {
		switch err {
		case storage.ErrAlreadyExists:
			return model.TokenPair{}, ErrAlreadyRegistered
		default:
			return model.TokenPair{}, fmt.Errorf("failed to register user: %w", err)
		}
	}

	tokenPair, err := s.GenerateTokenPair(userID)
	if err != nil {
		return model.TokenPair{}, err
	}

	return tokenPair, nil
}

func (s *service) Login(email string, password string) (model.TokenPair, error) {
	user, err := s.store.GetUser(email)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return model.TokenPair{}, ErrNotExist
		default:
			return model.TokenPair{}, err
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return model.TokenPair{}, ErrWrongPassword
	}

	tokenPair, err := s.GenerateTokenPair(user.ID)
	if err != nil {
		return model.TokenPair{}, err
	}

	return tokenPair, nil
}
