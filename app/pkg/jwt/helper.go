package jwt

import (
	"encoding/json"

	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"

	"time"

	"github.com/charopevez/eob-wayfinder-worker/internal/client/sentinel_service"
	"github.com/charopevez/eob-wayfinder-worker/internal/config"
	"github.com/charopevez/eob-wayfinder-worker/pkg/cache"
	"github.com/charopevez/eob-wayfinder-worker/pkg/logging"
)

var _ Helper = &helper{}

//?hash user email in jwt
type UserClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

type RT struct {
	RefreshToken string `json:"refresh_token"`
}

type helper struct {
	Logger  logging.Logger
	RTCache cache.Repository
}

func NewHelper(RTCache cache.Repository, logger logging.Logger) Helper {
	return &helper{RTCache: RTCache, Logger: logger}
}

type Helper interface {
	GenerateAccessToken(u sentinel_service.User) ([]byte, error)
	UpdateRefreshToken(rt RT) ([]byte, error)
}

func (h *helper) UpdateRefreshToken(rt RT) ([]byte, error) {
	defer h.RTCache.Del([]byte(rt.RefreshToken))

	userBytes, err := h.RTCache.Get([]byte(rt.RefreshToken))
	if err != nil {
		return nil, err
	}
	var u sentinel_service.User
	err = json.Unmarshal(userBytes, &u)
	if err != nil {
		return nil, err
	}
	return h.GenerateAccessToken(u)
}

func (h *helper) GenerateAccessToken(u sentinel_service.User) ([]byte, error) {
	key := []byte(config.GetConfig().JWT.Secret)
	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		return nil, err
	}
	builder := jwt.NewBuilder(signer)

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        u.UUID,
			Audience:  []string{"users"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 60)),
		},
		Email: u.Email,
	}
	token, err := builder.Build(claims)
	if err != nil {
		return nil, err
	}

	h.Logger.Info("create refresh token")
	refreshTokenUuid := uuid.New()
	userBytes, _ := json.Marshal(u)
	err = h.RTCache.Set([]byte(refreshTokenUuid.String()), userBytes, 0)
	if err != nil {
		h.Logger.Error(err)
		return nil, err
	}

	jsonBytes, err := json.Marshal(map[string]string{
		"token":         token.String(),
		"refresh_token": refreshTokenUuid.String(),
	})
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}
