package webhooks

import (
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/lucashthiele/chirpy/internal/config"
	"github.com/lucashthiele/chirpy/pkg/parser"
	"github.com/lucashthiele/chirpy/pkg/response"
)

type UpgradeUserParams struct {
	Event string `json:"event"`
	Data  struct {
		UserId uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func HandleUpgradeUser(res http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(res, err)
		return
	}

	params := &UpgradeUserParams{}

	err = parser.ParseBody(req.Body, params)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
		return
	}

	if params.Event != "user.upgraded" {
		response.RespondWithJSON(res, http.StatusNoContent, nil)
		return
	}

	err = cfg.Db.UpgradeUser(req.Context(), params.Data.UserId)
	if err == sql.ErrNoRows {
		response.RespondWithError(res, http.StatusNotFound, "Not found")
		return
	}
	if err != nil {
		response.RespondWithInternalServerError(res, err)
		return
	}

	response.RespondWithJSON(res, http.StatusNoContent, nil)
}
