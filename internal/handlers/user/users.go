package user

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lucashthiele/chirpy/internal/config"
	"github.com/lucashthiele/chirpy/pkg/parser"
	"github.com/lucashthiele/chirpy/pkg/response"
)

type params struct {
	Email string `json:"email"`
}

type userJSON struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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

	createdUser, err := cfg.Db.CreateUser(req.Context(), data.Email)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	userJSON := userJSON{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt.Time,
		UpdatedAt: createdUser.UpdatedAt.Time,
		Email:     createdUser.Email,
	}

	response.RespondWithJSON(res, http.StatusCreated, userJSON)
}
