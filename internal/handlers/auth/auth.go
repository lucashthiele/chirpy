package auth

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lucashthiele/chirpy/internal/auth"
	"github.com/lucashthiele/chirpy/internal/config"
	"github.com/lucashthiele/chirpy/internal/database"
	"github.com/lucashthiele/chirpy/pkg/parser"
	"github.com/lucashthiele/chirpy/pkg/response"
)

const expiresInOneHour time.Duration = time.Hour        // 1 hour
const expiresInDays time.Duration = time.Hour * 24 * 60 // 60 days

type params struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userJSON struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type tokenJSON struct {
	Token string `json:"token"`
}

func HandleLogin(resp http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	data := &params{}

	err = parser.ParseBody(req.Body, data)
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	user, err := cfg.Db.GetUserByEmail(req.Context(), data.Email)
	if err == sql.ErrNoRows {
		response.RespondWithError(resp, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	err = auth.CheckPassword(user.HashedPassword, data.Password)
	if err != nil {
		response.RespondWithError(resp, http.StatusUnauthorized, "Unauthorized")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.AppSecret, expiresInOneHour)
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	args := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(expiresInDays),
	}

	createdRefreshToken, err := cfg.Db.CreateRefreshToken(req.Context(), args)
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	userResp := userJSON{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt.Time,
		UpdatedAt:    user.UpdatedAt.Time,
		Email:        user.Email,
		Token:        token,
		RefreshToken: createdRefreshToken.Token,
		IsChirpyRed:  user.IsChirpyRed,
	}

	response.RespondWithJSON(resp, http.StatusOK, userResp)
}

func HandleRefreshToken(resp http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	refreshToken, err := auth.GetBearerToken(&req.Header)
	if err != nil {
		response.RespondWithError(resp, http.StatusUnauthorized, err.Error())
	}

	userId, err := cfg.Db.GetUserFromRefreshToken(req.Context(), refreshToken)
	if err == sql.ErrNoRows {
		response.RespondWithError(resp, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	token, err := auth.MakeJWT(userId, cfg.AppSecret, expiresInOneHour)
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	tokenResp := tokenJSON{
		Token: token,
	}

	response.RespondWithJSON(resp, http.StatusOK, tokenResp)
}

func HandleRevokeRefreshToken(resp http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	refreshToken, err := auth.GetBearerToken(&req.Header)
	if err != nil {
		response.RespondWithError(resp, http.StatusUnauthorized, err.Error())
	}

	err = cfg.Db.RevokeRefreshToken(req.Context(), refreshToken)
	if err != nil {
		response.RespondWithInternalServerError(resp, err)
		return
	}

	response.RespondWithJSON(resp, http.StatusNoContent, nil)
}
