package model

import (
	"github.com/dgrijalva/jwt-go"
)

type AccessToken struct {
	UserID    uint
	ExpiresAt int64
	jwt.StandardClaims
}

type RefreshToken struct {
	UserID    uint
	ExpiresAt int64
	jwt.StandardClaims
}

type TokenPair struct {
	AccessToken string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}