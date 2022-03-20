package storage

import "github.com/Nethius/tribble-customer-auth/pkg/model"

type Storage interface {
	GetUser(email string) (model.User, error)
	InsertUser(email string, password string) (uint, error)
	GetDeviceList(userID uint) ([]string, error)
	InsertDevice(userID uint) error
	//UpdateAccount(email string, password string) (model.User, error)
	//DeleteAccount(email string, password string) (model.User, error)
}
