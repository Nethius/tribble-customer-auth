package register

import (
	"encoding/json"
	"errors"
	"github.com/Nethius/tribble-customer-auth/pkg/model"
	"github.com/Nethius/tribble-customer-auth/pkg/service/auth"
	"github.com/rs/zerolog"
	"net/http"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInternal       = errors.New("something went wrong")
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
	service auth.Service
	logger  zerolog.Logger
}

func NewHandler(service auth.Service, logger zerolog.Logger) *Handler {
	l := logger.With().Str("component", "registerHandler").Logger()
	return &Handler{service: service, logger: l}
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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	user := &model.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		h.logger.Error().Msgf("failed to decode request %v: %v", r.Body, err)
		h.respond(w, errorMessage{Error: ErrInvalidRequest.Error()}, http.StatusBadRequest)
		return
	}

	tokens, err := h.service.Register(user.Email, user.Password)
	if err != nil {
		h.logger.Error().Msgf("failed to register user %v: %v", user.Email, err)
		switch err {
		case auth.ErrAlreadyRegistered:
			h.respond(w, errorMessage{Error: auth.ErrAlreadyRegistered.Error()}, http.StatusForbidden)
			return
		default:
			h.respond(w, errorMessage{Error: ErrInternal.Error()}, http.StatusInternalServerError)
			return
		}
	}

	resp := response{Message: "user successfully created", AccessToken: tokens.AccessToken,
		RefreshToken: tokens.RefreshToken}
	h.respond(w, resp, http.StatusCreated)
}
