package jwt

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Nethius/tribble-customer-auth/pkg/model"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrMissingAccessToken       = errors.New("missing access token")
	ErrInvalidAccessToken       = errors.New("invalid access token")
	ErrMalformedAccessToken     = errors.New("malformed access token")
	ErrExpiredAccessToken       = errors.New("expired access token")
	ErrMissingAccessTokenSecret = errors.New("missing access token secret")
)

type errorMessage struct {
	Error string `json:"errorMessage"`
}

type middleware struct {
	logger       zerolog.Logger
	accessSecret string
}

func NewMiddleware(logger zerolog.Logger) *middleware {
	l := logger.With().Str("component", "middleware").Logger()
	m := middleware{logger: l}
	m.setupAccessTokenSecret()
	return &m
}

func (m *middleware) setupAccessTokenSecret() {
	m.accessSecret = os.Getenv("ACCESSSECRET")
}

func (m *middleware) respond(w http.ResponseWriter, data interface{}, code int) {
	w.WriteHeader(code)
	if data != nil {
		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			m.logger.Error().Msgf("failed to write response: %v", err)
		}
	}
}

func (m *middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.accessSecret == "" {
			m.logger.Error().Msgf("%v", ErrMissingAccessTokenSecret)
			m.respond(w, errorMessage{Error: ErrMissingAccessTokenSecret.Error()}, http.StatusInternalServerError)
			return
		}

		notAuth := []string{"/api/user/register", "/api/user/login", "/api/user/refreshToken"}
		requestPath := r.URL.Path

		for _, value := range notAuth {
			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		tokenHeader := r.Header.Get("authorization")

		if (tokenHeader == "Bearer") || (tokenHeader == "") {
			m.logger.Info().Msgf("%v", ErrMissingAccessToken)
			m.respond(w, errorMessage{Error: ErrMissingAccessToken.Error()}, http.StatusUnauthorized)
			return
		}

		chunks := strings.Split(tokenHeader, " ") // should look like "Bearer token"
		if len(chunks) != 2 && chunks[0] != "Bearer " {
			m.logger.Info().Msgf("missing token part: %v", ErrInvalidAccessToken)
			m.respond(w, errorMessage{Error: ErrInvalidAccessToken.Error()}, http.StatusForbidden)
			return
		}

		token := chunks[1]

		tk := &model.AccessToken{}

		accessToken, err := jwt.ParseWithClaims(token, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.accessSecret), nil
		})

		if err != nil {
			m.logger.Info().Msgf("%v", ErrMalformedAccessToken)
			m.respond(w, errorMessage{Error: ErrMalformedAccessToken.Error()}, http.StatusForbidden)
			return
		}

		if !accessToken.Valid {
			m.logger.Info().Msgf("token in not valid: %v", ErrInvalidAccessToken)
			m.respond(w, errorMessage{Error: ErrInvalidAccessToken.Error()}, http.StatusForbidden)
			return
		}

		if tk.ExpiresAt < time.Now().Unix() {
			m.logger.Info().Msgf("%v", ErrExpiredAccessToken)
			m.respond(w, errorMessage{Error: ErrExpiredAccessToken.Error()}, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", tk.UserID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
