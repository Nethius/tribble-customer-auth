package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Nethius/tribble-customer-auth/pkg/handlers/login"
	"github.com/Nethius/tribble-customer-auth/pkg/handlers/refreshtoken"
	"github.com/Nethius/tribble-customer-auth/pkg/handlers/register"
	"github.com/Nethius/tribble-customer-auth/pkg/middleware/jwt"
	"github.com/Nethius/tribble-customer-auth/pkg/service/auth"
	"github.com/Nethius/tribble-customer-auth/pkg/storage/postgres"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"log"
	"net/http"
	"os"
)

func main() {
	mainLogger := initLogger()
	logger := mainLogger.With().Str("component", "Main").Logger()

	err := godotenv.Load()
	if err != nil {
		logger.Panic().Msg(err.Error())
	}

	connStr, err := getPostgresCredentials()
	if err != nil {
		logger.Panic().Msgf("failed to get db credentials from env: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Panic().Msgf("failed to open db connection: %v", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		logger.Panic().Msgf("failed to open db connection: %v", err)
	}

	defer db.Close()

	psDB := postgres.NewPostgres(db)

	authService := auth.NewService(psDB)

	router := mux.NewRouter()

	m := jwt.NewMiddleware(mainLogger)
	router.Use(m.Handler)

	registerHandler := register.NewHandler(authService, mainLogger)
	loginHandler := login.NewHandler(authService, mainLogger)
	refreshTokenHandler := refreshtoken.NewHandler(authService, mainLogger)

	router.HandleFunc("/api/user/register", registerHandler.Register)
	router.HandleFunc("/api/user/login", loginHandler.Login)
	router.HandleFunc("/api/user/refreshToken", refreshTokenHandler.RefreshToken)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func initLogger() zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func getPostgresCredentials() (string, error) {
	host, ok := os.LookupEnv("PGHOST")
	if !ok {
		return "", errors.New("failed to get PGHOST from env")
	}

	port, ok := os.LookupEnv("PGPORT")
	if !ok {
		return "", errors.New("failed to get PGPORT from env")
	}

	user, ok := os.LookupEnv("PGUSER")
	if !ok {
		return "", errors.New("failed to get PGUSER from env")
	}

	password, ok := os.LookupEnv("PGPASSWORD")
	if !ok {
		return "", errors.New("failed to get PGPASSWORD from env")
	}

	dbname, ok := os.LookupEnv("PGDATABASE")
	if !ok {
		return "", errors.New("failed to get PGDATABASE from env")
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname), nil
}
