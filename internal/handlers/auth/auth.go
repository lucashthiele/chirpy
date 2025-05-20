package auth

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lucashthiele/chirpy/internal/auth"
	"github.com/lucashthiele/chirpy/internal/config"
	"github.com/lucashthiele/chirpy/pkg/parser"
	"github.com/lucashthiele/chirpy/pkg/response"
)

const maximumExpiresInSeconds int = 60 * 60

type params struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type userJSON struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func HandleLogin(res http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(res, err)
		return
	}

	data := &params{}

	err = parser.ParseBody(req.Body, data)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
		return
	}

	if data.ExpiresInSeconds <= 0 ||
		data.ExpiresInSeconds > maximumExpiresInSeconds {
		data.ExpiresInSeconds = maximumExpiresInSeconds
	}

	user, err := cfg.Db.GetUserByEmail(req.Context(), data.Email)
	if err == sql.ErrNoRows {
		response.RespondWithError(res, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err != nil {
		response.RespondWithInternalServerError(res, err)
		return
	}

	err = auth.CheckPassword(user.HashedPassword, data.Password)
	if err != nil {
		response.RespondWithError(res, http.StatusUnauthorized, "Unauthorized")
		return
	}

	duration := time.Duration(time.Second * time.Duration(data.ExpiresInSeconds))

	token, err := auth.MakeJWT(user.ID, cfg.AppSecret, duration)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
		return
	}

	userResp := userJSON{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email,
		Token:     token,
	}

	response.RespondWithJSON(res, http.StatusOK, userResp)
}
