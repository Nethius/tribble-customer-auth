package postgres

import (
	"database/sql"
	"github.com/Nethius/tribble-customer-auth/pkg/model"
	"github.com/Nethius/tribble-customer-auth/pkg/storage"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) GetUser(email string) (model.User, error) {
	q := `SELECT id, email, password FROM users WHERE email = $1`

	user := model.User{}
	row := p.db.QueryRow(q, email)
	if err := row.Scan(&user.ID, &user.Email, &user.Password); err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (p *Postgres) InsertUser(email string, password string) (uint, error) {
	q := `INSERT INTO users (email, password) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id`

	var userId uint = 0
	if err := p.db.QueryRow(q, email, password).Scan(&userId); err != nil {
		switch err {
		case sql.ErrNoRows:
			return 0, storage.ErrAlreadyExists
		default:
			return 0, err
		}
	}

	return userId, nil
}

func (p *Postgres) GetDeviceList(userID uint) ([]string, error) {
	q := `SELECT imei FROM devices WHERE user_id = $1`

	var deviceList []string

	rows, err := p.db.Query(q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var imei string
		if err := rows.Scan(&imei); err != nil {
			return nil, err
		}
		deviceList = append(deviceList, imei)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deviceList, nil
}

func (p *Postgres) InsertDevice(userID uint) error {
	return nil
}

//func (p *Postgres) UpdateAccount(email string, password string) (model.User, error) {
//
//}
//
//func (p *Postgres) DeleteAccount(email string, password string) (model.User, error) {
//
//}
