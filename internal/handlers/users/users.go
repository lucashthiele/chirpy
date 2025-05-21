package users

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lucashthiele/chirpy/internal/auth"
	"github.com/lucashthiele/chirpy/internal/config"
	"github.com/lucashthiele/chirpy/internal/database"
	"github.com/lucashthiele/chirpy/pkg/parser"
	"github.com/lucashthiele/chirpy/pkg/response"
)

type params struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userJSON struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func HandleCreateUsers(res http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	data := &params{}

	err = parser.ParseBody(req.Body, data)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	hashedPassword, err := auth.HashPassword(data.Password)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	userParams := database.CreateUserParams{
		Email:          data.Email,
		HashedPassword: hashedPassword,
	}

	createdUser, err := cfg.Db.CreateUser(req.Context(), userParams)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	userJSON := userJSON{
		ID:          createdUser.ID,
		CreatedAt:   createdUser.CreatedAt.Time,
		UpdatedAt:   createdUser.UpdatedAt.Time,
		Email:       createdUser.Email,
		IsChirpyRed: createdUser.IsChirpyRed,
	}

	response.RespondWithJSON(res, http.StatusCreated, userJSON)
}

func HandleUpdateUsers(res http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	data := &params{}

	err = parser.ParseBody(req.Body, data)
	if err != nil {
		response.RespondWithError(res, http.StatusBadRequest, err.Error())
	}

	userId, ok := req.Context().Value(config.UserIDKey).(uuid.UUID)
	if !ok {
		response.RespondWithInternalServerError(res, fmt.Errorf("omg you're so bad at this"))
		return
	}

	hashedPassword, err := auth.HashPassword(data.Password)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	updateParams := database.UpdateUserEmailAndPasswordParams{
		ID:             userId,
		Email:          data.Email,
		HashedPassword: hashedPassword,
	}

	updatedUser, err := cfg.Db.UpdateUserEmailAndPassword(req.Context(), updateParams)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	userJSON := userJSON{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt.Time,
		UpdatedAt:   updatedUser.UpdatedAt.Time,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	}

	response.RespondWithJSON(res, http.StatusOK, userJSON)
}
