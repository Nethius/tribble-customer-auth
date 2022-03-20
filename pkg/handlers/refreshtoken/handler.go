package refreshtoken

import (
	"encoding/json"
	"errors"
	"github.com/Nethius/tribble-customer-auth/pkg/model"
	"github.com/Nethius/tribble-customer-auth/pkg/service/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrMissingRefreshToken       = errors.New("missing refresh token")
	ErrInvalidRefreshToken       = errors.New("invalid refresh token")
	ErrMalformedRefreshToken     = errors.New("malformed refresh token")
	ErrExpiredRefreshToken       = errors.New("expired Refresh token")
	ErrMissingRefreshTokenSecret = errors.New("missing refresh token secret")
)

type response struct {
	Message      string `json:"message"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type errorMessage struct {
	Error string `json:"errorMessage"`
}

type Handler struct {
	service       auth.Service
	logger        zerolog.Logger
	refreshSecret string
}

func NewHandler(service auth.Service, logger zerolog.Logger) *Handler {
	l := logger.With().Str("component", "refreshTokenHandler").Logger()
	h := Handler{service: service, logger: l}
	h.setupRefreshTokenSecret()
	return &h
}

func (h *Handler) setupRefreshTokenSecret() {
	h.refreshSecret = os.Getenv("REFRESHSECRET")
}

func (h *Handler) respond(w http.ResponseWriter, data interface{}, code int) {
	w.WriteHeader(code)
	if data != nil {
		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error().Msgf("failed to write response: %v", err)
		}
	}
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if h.refreshSecret == "" {
		h.logger.Error().Msgf("%v", ErrMissingRefreshTokenSecret)
		h.respond(w, errorMessage{Error: ErrMissingRefreshTokenSecret.Error()}, http.StatusInternalServerError)
		return
	}

	tokenHeader := r.Header.Get("authorization")

	if (tokenHeader == "Bearer") || (tokenHeader == "") {
		h.logger.Info().Msgf("%v", ErrMissingRefreshToken)
		h.respond(w, errorMessage{Error: ErrMissingRefreshToken.Error()}, http.StatusUnauthorized)
		return
	}

	chunks := strings.Split(tokenHeader, " ") // should look like "Bearer token"
	if len(chunks) != 2 && chunks[0] != "Bearer " {
		h.logger.Info().Msgf("missing token part: %v", ErrInvalidRefreshToken)
		h.respond(w, errorMessage{Error: ErrInvalidRefreshToken.Error()}, http.StatusForbidden)
		return
	}

	token := chunks[1] //skip "Bearer " part of message

	tk := &model.RefreshToken{}

	refreshToken, err := jwt.ParseWithClaims(token, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.refreshSecret), nil
	})

	if err != nil {
		h.logger.Info().Msgf("%v", ErrMalformedRefreshToken)
		h.respond(w, errorMessage{Error: ErrMalformedRefreshToken.Error()}, http.StatusForbidden)
		return
	}

	if !refreshToken.Valid {
		h.logger.Info().Msgf("token in not valid: %v", ErrInvalidRefreshToken)
		h.respond(w, errorMessage{Error: ErrInvalidRefreshToken.Error()}, http.StatusForbidden)
		return
	}

	if tk.ExpiresAt < time.Now().Unix() {
		h.logger.Info().Msgf("%v", ErrExpiredRefreshToken)
		h.respond(w, errorMessage{Error: ErrExpiredRefreshToken.Error()}, http.StatusUnauthorized)
		return
	}

	tokens, err := h.service.GenerateTokenPair(tk.UserID)
	if err != nil {
		h.logger.Error().Msgf("failed to generate tokens: %v", err)
		h.respond(w, errorMessage{Error: err.Error()}, http.StatusInternalServerError)
		return
	}

	resp := response{Message: "tokens successfully refreshed", AccessToken: tokens.AccessToken,
		RefreshToken: tokens.RefreshToken}
	h.respond(w, resp, http.StatusCreated)
}
