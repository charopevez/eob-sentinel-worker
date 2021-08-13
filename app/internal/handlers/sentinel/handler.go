package sentinel

import (
	"encoding/json"

	"github.com/julienschmidt/httprouter"

	"net/http"

	"github.com/charopevez/eob-wayfinder-worker/internal/apperror"
	"github.com/charopevez/eob-wayfinder-worker/internal/client/sentinel_service"
	"github.com/charopevez/eob-wayfinder-worker/pkg/jwt"
	"github.com/charopevez/eob-wayfinder-worker/pkg/logging"
)

const (
	signUpURL       = "/api/register"
	getTokenURL     = "/api/token"
	refreshTokenURL = "/api/token/refresh"
	signOutURL      = "/api/logout"
)

type Handler struct {
	Logger          logging.Logger
	SentinelService sentinel_service.SentinelService
	JWTHelper       jwt.Helper
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, signUpURL, apperror.Middleware(h.SignUp))
	router.HandlerFunc(http.MethodPost, getTokenURL, apperror.Middleware(h.GetToken))
	router.HandlerFunc(http.MethodPost, refreshTokenURL, apperror.Middleware(h.RefreshToken))
	router.HandlerFunc(http.MethodGet, signOutURL, apperror.Middleware(h.LogOut))
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	var dto sentinel_service.CreateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return apperror.BadRequestError("failed to decode data")
	}

	u, err := h.SentinelService.Create(r.Context(), dto)
	if err != nil {
		return err
	}
	token, err := h.JWTHelper.GenerateAccessToken(u)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(token)

	return nil
}

func (h *Handler) GetToken(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	var token []byte
	var err error
	defer r.Body.Close()
	var dto sentinel_service.SignInUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return apperror.BadRequestError("failed to decode data")
	}
	u, err := h.SentinelService.GetByEmailAndPassword(r.Context(), dto.Email, dto.Password)
	if err != nil {
		return err
	}
	token, err = h.JWTHelper.GenerateAccessToken(u)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(token)

	return err
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	var token []byte
	var err error
	defer r.Body.Close()
	var rt jwt.RT
	if err := json.NewDecoder(r.Body).Decode(&rt); err != nil {
		return apperror.BadRequestError("failed to decode data")
	}
	token, err = h.JWTHelper.UpdateRefreshToken(rt)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(token)

	return err
}

func (h *Handler) LogOut(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	var token []byte
	var err error
	defer r.Body.Close()
	// var rt jwt.RT
	// if err := json.NewDecoder(r.Body).Decode(&rt); err != nil {
	// 	return apperror.BadRequestError("failed to decode data")
	// }
	// token, err = h.JWTHelper.UpdateRefreshToken(rt)
	// if err != nil {
	// 	return err
	// }

	w.WriteHeader(http.StatusCreated)
	w.Write(token)

	return err
}
