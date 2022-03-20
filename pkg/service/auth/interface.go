package auth

import "github.com/Nethius/tribble-customer-auth/pkg/model"

type Service interface {
	GenerateTokenPair(userID uint) (model.TokenPair, error)

	Register(email string, password string) (model.TokenPair, error)
	Login(email string, password string) (model.TokenPair, error)
}
